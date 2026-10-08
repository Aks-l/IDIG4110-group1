package simulator

import (
	"math/rand"
)

const (
	//min, max, and max change values for the simulated temperature sensor
	MinTemperature = 10.0
	MaxTemperature = 50.0
	MaxTempChange  = 0.3
)

package simulator

import "math/rand"

const (
	minTemperature = 16.0
	maxTemperature = 28.0
	targetTemperature = 21.0

	// How strongly the temperature moves back toward the target.
	driftStrength = 0.05

	// Maximum random sensor/environment variation each reading.
	randomVariation = 0.2
)

func NextTemperature(current float64) float64 {
	// Slowly move toward normal room temperature.
	drift := (targetTemperature - current) * driftStrength

	// Add random environmental/sensor variation.
	noise := (rand.Float64()*2 - 1) * randomVariation

	next := current + drift + noise

	if next < minTemperature {
		return minTemperature
	}

	if next > maxTemperature {
		return maxTemperature
	}

	return next
}
