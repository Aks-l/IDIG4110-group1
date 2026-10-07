package simulator

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"IDIG4110/shared/dto"
)

// Builds one HA state_changed event for a simulated temperature entity
//
// # Inputs:
//
//   - entityID [string] simulated entity id
//
// # Returns:
//
//   - Sensor state event
func GenerateMockData(entityID string) dto.SensorStateEvent {
	now := time.Now()
	state := strconv.FormatFloat(18+rand.Float64()*8, 'f', 1, 64)
	contextID := newContextID()

	return dto.SensorStateEvent{
		EventType: "state_changed",
		TimeFired: now,
		Data: dto.EventData{
			EntityID: entityID,
			NewState: dto.StateObject{
				EntityID: entityID,
				State:    state,
				Attributes: map[string]any{
					"device_class":        "temperature",
					"unit_of_measurement": "°C",
					"friendly_name":       friendlyName(entityID),
				},
				LastChanged: now,
				LastUpdated: now,
				Context:     dto.EventContext{ID: contextID},
			},
		},
		Context: dto.EventContext{ID: contextID},
	}
}

// Builds HA style context id, 32 hex characters
//
// # Returns:
//
//   - Random context id
func newContextID() string {
	return fmt.Sprintf("%016x%016x", rand.Uint64(), rand.Uint64())
}

func friendlyName(entityID string) string {
	return strings.ReplaceAll(strings.TrimPrefix(entityID, "sensor."), "_", " ")
}
