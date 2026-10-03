package mqttclient

import (
	"IDIG4110/shared/dto"
	"encoding/json"
	"log/slog"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)


type Collector struct {
	workers []chan dto.SensorStateEvent
}

func NewCollector(workerCount int, bufferSize int) *Collector {
	w := make([]chan dto.SensorStateEvent, workerCount)

	for i := range w {
		w[i] = make(chan dto.SensorStateEvent, bufferSize)
	}

	return &Collector{
		workers: w,
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
	case c.workers[workerIndex] <-state:
	default:
		slog.Warn("MQTT Collection is full")
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
