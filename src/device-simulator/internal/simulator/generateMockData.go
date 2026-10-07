package simulator

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"IDIG4110/shared/dto"
)

const (
	minTemperature = 18.0
	maxTemperature = 24.0
)

// NextTemperature performs a bounded random walk from the current temperature
func NextTemperature(current float64) float64 {
	next := current + (rand.Float64()-0.5)*0.8
	if next < minTemperature {
		next = minTemperature
	}
	if next > maxTemperature {
		next = maxTemperature
	}
	return next
}

// GenerateMockData generates a device-native state event with a string state
func GenerateMockData(entityID string, temperature float64) dto.StateEvent {
	return dto.StateEvent{
		EventType: "state_changed",
		TimeFired: time.Now(),
		EntityID:  entityID,
		NewState: dto.NewState{
			State: strconv.FormatFloat(temperature, 'f', 1, 64),
			Attributes: map[string]string{
				"device_class":        "temperature",
				"unit_of_measurement": "°C",
				"friendly_name":       fmt.Sprintf("Simulated temperature sensor %s", entityID),
			},
		},
	}
}
