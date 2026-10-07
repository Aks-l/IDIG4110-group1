package mqttclient

import (
	"context"
	"encoding/json"
	"log/slog"

	"IDIG4110/ingest-service/internal/domain"
	"IDIG4110/shared/dto"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Collector struct {
	workers []chan dto.RawMessage
	svc     domain.SensorIngestSvc
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
		slog.Error("MQTT Handler", "error", err, "topic", msg.Topic())
		return
	}

	if err := c.svc.CreateRaw(context.Background(), raw); err != nil {
		slog.Error("Failed to store raw message", "error", err, "topic", raw.Topic)
	}

	workerIndex := hash(raw.Topic) % len(c.workers)

	select {
	case c.workers[workerIndex] <- raw:
	default:
		slog.Warn("MQTT Collection is full")
	}
}

func (c *Collector) StartWorkers() {
	for i, ch := range c.workers {
		go func(entityID int, queue chan dto.RawMessage) {
			for raw := range queue {
				if err := c.svc.Create(context.Background(), raw); err != nil {
					slog.Error("Failed to create sensor object", "error", err)
				}
			}
		}(i, ch)
	}
}

func (c *Collector) Close() {
	for i := range c.workers {
		close(c.workers[i])
	}
}

func hash(s string) int {
	h := 0
	for _, c := range s {
		h += int(c)
	}
	return h
}
