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
	workers []chan dto.SensorStateEvent
	svc     domain.SensorIngestSvc
	wg      sync.WaitGroup
}

func NewCollector(workerCount, bufferSize int, svc domain.SensorIngestSvc) *Collector {
	w := make([]chan dto.SensorStateEvent, workerCount)

	for i := range w {
		w[i] = make(chan dto.SensorStateEvent, bufferSize)
	}

	return &Collector{
		workers: w,
		svc:     svc,
	}
}

func (c *Collector) MQTTHandler(client mqtt.Client, msg mqtt.Message) {
	var state dto.SensorStateEvent

	if err := json.Unmarshal(msg.Payload(), &state); err != nil {
		slog.Error("MQTT Handler", "error", err)
		return
	}
	workerIndex := hash(state.EntityID) % len(c.workers)

	select {
	case c.workers[workerIndex] <- state:
	default:
		slog.Warn("MQTT Collection is full")
	}
}

func (c *Collector) StartWorkers() {
	for i, ch := range c.workers {
		c.wg.Add(1)

		go func(entityID int, queue chan dto.SensorStateEvent) {
			defer c.wg.Done()

			for state := range queue {
				if err := c.svc.Create(context.Background(), state); err != nil {
					slog.Error(
						"Failed to create sensor object",
						"entityId", state.EntityID,
						"error", err)
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
