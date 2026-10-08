package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"IDIG4110/shared/dto"
)

// Normalize converts a raw gateway message into a normalized reading, per
// docs/architecture/mqtt-envelope.md. Identity comes from the event payload:
// external_entity_id is the source's own entity id, never the MQTT topic.
// Topics are transport routing; identity flows end to end as the
// (gateway_id, external_entity_id) pair.
func Normalize(raw dto.RawMessage) (dto.Reading, error) {
	var event dto.StateEvent
	if err := json.Unmarshal(raw.Payload, &event); err != nil {
		return dto.Reading{}, fmt.Errorf("normalize: unmarshal payload: %w", err)
	}

	entityID := strings.TrimSpace(event.EntityID)
	if entityID == "" {
		return dto.Reading{}, fmt.Errorf("normalize: payload has no entity_id")
	}

	reading := dto.Reading{
		GatewayID:        raw.GatewayID,
		ExternalEntityID: entityID,
		Timestamp:        eventTime(event, raw),
		DeviceClass:      attrString(event.NewState.Attributes, "device_class"),
		Unit:             attrString(event.NewState.Attributes, "unit_of_measurement"),
		Attributes:       map[string]any{},
	}

	for key, value := range event.NewState.Attributes {
		if key == "device_class" || key == "unit_of_measurement" {
			continue
		}
		reading.Attributes[key] = value
	}
	if len(reading.Attributes) == 0 {
		reading.Attributes = nil
	}

	// States are always strings at the source. Numeric ones become value_num;
	// everything else (modes, on/off, and sentinels such as unavailable and
	// unknown, which carry meaning) stays verbatim in value_text.
	if valueNum, err := strconv.ParseFloat(strings.TrimSpace(event.NewState.State), 64); err == nil {
		reading.ValueNum = &valueNum
	} else {
		state := event.NewState.State
		reading.ValueText = &state
	}

	return reading, nil
}

// eventTime prefers the event's own time_fired; the envelope time is only a
// fallback for sources that do not stamp their events.
func eventTime(event dto.StateEvent, raw dto.RawMessage) time.Time {
	if !event.TimeFired.IsZero() {
		return event.TimeFired
	}
	return raw.Time
}

// attrString returns a string attribute, nil for absent or non-string values
func attrString(attributes map[string]any, key string) *string {
	if value, ok := attributes[key].(string); ok {
		return &value
	}
	return nil
}
