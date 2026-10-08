package simulator

import (
	"fmt"
	"time"

	"IDIG4110/shared/dto"
)

// GenerateMockData generates a device-native state event with a string state
func GenerateMockData(entityID string, temperature float64) dto.StateEvent {
	return dto.StateEvent{
		EventType: "state_changed",
		TimeFired: time.Now(),
		EntityID:  entityID,
		NewState: dto.NewState{
			State: fmt.Sprintf("%.1f", temperature),
			Attributes: map[string]string{
				"device_class":        "temperature",
				"unit_of_measurement": "°C",
				"friendly_name":       fmt.Sprintf("Simulated temperature sensor %s", entityID),
			},
		},
	}
}
