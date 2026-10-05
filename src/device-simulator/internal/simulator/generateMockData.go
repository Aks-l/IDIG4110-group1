package simulator

import (
	"math/rand"
	"strconv"
	"time"

	"IDIG4110/shared/dto"
)

func GenerateMockData(eventId string) dto.SensorStateEvent {
	return dto.SensorStateEvent{
		EventType: "test",
		TimeFired: time.Now(),
		EntityID:  eventId,
		NewState: dto.NewState{
			State: strconv.Itoa(int(rand.Int()) * 20),
			Attributes: map[string]string{
				"device_class":        "temperature",
				"unit_of_measurement": "°C",
				"friendly_name":       "Living room temperature",
			},
		},
	}
}
