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

func NextTemperature(current float64) float64 {

	//change temperature between -0.3 and +0.3 degrees Celcius
	change := rand.Float64()*(2*MaxTempChange) - MaxTempChange
	next := current + change

	//keep the simulated temperature inside the min and max range
	if next < MinTemperature {
		next = MinTemperature
	}

	if next > MaxTemperature {
		return MaxTemperature
	}
	return next
}
