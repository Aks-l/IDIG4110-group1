package mqttclient

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"IDIG4110/ingest-service/internal/domain"
	"IDIG4110/shared/dto"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Collector struct {
	workers []chan dto.RawMessage
	svc     domain.SensorIngestSvc
	wg      sync.WaitGroup
}

func NewCollector(workerCount, bufferSize int, svc domain.SensorIngestSvc) *Collector {
	w := make([]chan dto.RawMessage, workerCount)

	for i := range w {
		w[i] = make(chan dto.RawMessage, bufferSize)
	}

	return &Collector{
		workers: w,
		svc:     svc,
	}
}

func (c *Collector) MQTTHandler(client mqtt.Client, msg mqtt.Message) {
	var raw dto.RawMessage

	if err := json.Unmarshal(msg.Payload(), &raw); err != nil {
		slog.Error(
			"failed to decode MQTT sensor event",
			"topic", msg.Topic(),
			"error", err,
		)
		return
	}

	if err := c.svc.CreateRaw(context.Background(), raw); err != nil {
		slog.Error("Failed to store raw message", "error", err, "topic", raw.Topic)
	}

	// The topic only routes the message to a worker for per-entity ordering;
	// the sensor's identity lives in the payload (see service.Normalize).
	workerIndex := hash(raw.Topic) % len(c.workers)

	select {
	case c.workers[workerIndex] <- raw:
	default:
		slog.Warn(
			"MQTT Collection is full, dropping event",
			"topic", raw.Topic,
			"worker", workerIndex)
	}
}

func (c *Collector) StartWorkers() {
	for _, ch := range c.workers {

		c.wg.Add(1)

		go func(queue chan dto.RawMessage) {
			defer c.wg.Done()

			for raw := range queue {
				if err := c.svc.Create(context.Background(), raw); err != nil {
					slog.Error(
						"Failed to create sensor object",
						"topic", raw.Topic,
						"error", err)
				}
			}
		}(ch)
	}
}

func (c *Collector) Close() {
	for _, worker := range c.workers {
		close(worker)
	}
	c.wg.Wait()
}

func hash(s string) int {
	h := 0
	for _, c := range s {
		h += int(c)
	}
	return h
}
