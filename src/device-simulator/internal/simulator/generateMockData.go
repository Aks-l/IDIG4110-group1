package simulator

import (
	"fmt"
	"math/rand"
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

func GenerateMockData(entityID, gatewayID string, temperature float64) dto.Reading {
	return dto.Reading{
		GatewayID:        gatewayID,
		ExternalEntityID: entityID,
		Time:             time.Now(),
		DeviceClass:      "temperature",
		ValueNum:         &temperature,
		Unit:             "°C",
		Attributes: map[string]any{
			"friendly_name": fmt.Sprintf("Simulated temperature sensor %s", entityID),
		},
	}
}
