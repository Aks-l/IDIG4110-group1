package simulator

import (
	"IDIG4110/shared/dto"
	"crypto/rand"
	"fmt"
	"time"
)

type TemperatureSensor struct {
	ID      string
	Current float64
	Min     float64
	Max     float64
}

func NewTemperatureSensor(id string) *TemperatureSensor {
	return &TemperatureSensor{
		ID:      id,
		Current: 20.0 + rand.Float64()*0.3, // Random initial temperature between min and max
		Min:     10,
		Max:     50,
	}
}

func (s *TemperatureSensor) Generate() dto.SensorStateEvent {
	//change temperature between -0.3 and +0.3 degrees Celcius
	change := rand.Float64()*0.6 - 0.3
	s.Current += change

	//keep the simulated temperature inside the min and max range
	if s.Current < s.Min {
		s.Current = s.Min
	}

	if s.Current > s.Max {
		s.Current = s.Max
	}

	return dto.SensorStateEvent{
		EventType: "temperature",
		TimeFired: time.Now().UTC(),
		EntityID:  s.ID,
		NewState: dto.NewState{
			State: fmt.Sprint("%.1f", s.Current),
			Attributes: map[string]string{
				"device_class":        "temperature",
				"unit_of_measurement": "°C",
				"friendly_name":       "Living Room Temperature",
			},
		},
	}
}
