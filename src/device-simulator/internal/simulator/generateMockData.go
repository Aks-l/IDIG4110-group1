package simulator

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"IDIG4110/shared/dto"
)

// GenerateMockData builds one HA-style state_changed event for a simulated
// temperature entity, as the HA Green hub would emit it.
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

// newContextID mimics HA context ids: 32 hex characters.
func newContextID() string {
	return fmt.Sprintf("%016x%016x", rand.Uint64(), rand.Uint64())
}

func friendlyName(entityID string) string {
	return strings.ReplaceAll(strings.TrimPrefix(entityID, "sensor."), "_", " ")
}
