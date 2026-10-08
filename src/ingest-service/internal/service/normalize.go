package service

import (
	"encoding/json"
	"fmt"
	"strconv"

	"IDIG4110/shared/dto"
)

// Normalize converts a raw gateway message into a normalized reading
func Normalize(raw dto.RawMessage) (dto.Reading, error) {
	var event dto.StateEvent
	if err := json.Unmarshal(raw.Payload, &event); err != nil {
		return dto.Reading{}, fmt.Errorf("normalize: unmarshal payload: %w", err)
	}

	reading := dto.Reading{
		GatewayID:        raw.GatewayID,
		ExternalEntityID: event.EntityID,
		EventID:          &event.EntityID,
		Timestamp:        raw.Time,
		DeviceClass:      attrPtr(event.NewState.Attributes, "device_class"),
		Unit:             attrPtr(event.NewState.Attributes, "unit_of_measurement"),
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

	if valueNum, err := strconv.ParseFloat(event.NewState.State, 64); err == nil {
		reading.ValueNum = &valueNum
	} else {
		state := event.NewState.State
		reading.ValueText = &state
	}

	return reading, nil
}

func attrPtr(attributes map[string]string, key string) *string {
	if value, ok := attributes[key]; ok {
		return &value
	}
	return nil
}
