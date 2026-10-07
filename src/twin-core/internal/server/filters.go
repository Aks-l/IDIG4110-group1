package server

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"IDIG4110/shared/dto"
	"IDIG4110/twin-core/internal/domain"
)

const (
	defaultEntityPageLimit = 100
	maxEntityPageLimit     = 500
)

// Parses and validates /state list query params
// unknown keys, repeats and malformed values fail with ErrBadRequest
//
// # Inputs:
//
//   - q [url.Values] query params
//
// # Returns:
//
//   - Entity filter, limit defaults to 100
//   - ErrBadRequest on any invalid param
func parseEntityFilter(q url.Values) (domain.EntityFilter, error) {
	filter := domain.EntityFilter{Limit: defaultEntityPageLimit}

	for key, values := range q {
		if len(values) != 1 {
			return filter, fmt.Errorf("%w: repeated query param %q", domain.ErrBadRequest, key)
		}
		value := values[0]

		var err error
		switch key {
		case "home_id":
			filter.HomeID, err = uuidParam(key, value)
		case "area_id":
			filter.AreaID, err = uuidParam(key, value)
		case "device_id":
			filter.DeviceID, err = uuidParam(key, value)
		case "domain":
			filter.Domains, err = listParam(key, value)
		case "device_class":
			filter.DeviceClasses, err = listParam(key, value)
		case "device_type":
			filter.DeviceTypes, err = listParam(key, value)
		case "controllable":
			filter.Controllable, err = boolParam(key, value)
		case "floor":
			var floor int
			floor, err = intParam(key, value)
			filter.Floor = &floor
		case "limit":
			filter.Limit, err = intParam(key, value)
			if err == nil && (filter.Limit < 1 || filter.Limit > maxEntityPageLimit) {
				return filter, fmt.Errorf("%w: limit must be between 1 and %d", domain.ErrBadRequest, maxEntityPageLimit)
			}
		case "offset":
			filter.Offset, err = intParam(key, value)
			if err == nil && filter.Offset < 0 {
				return filter, fmt.Errorf("%w: offset must be 0 or greater", domain.ErrBadRequest)
			}
		default:
			return filter, fmt.Errorf("%w: unknown query param %q", domain.ErrBadRequest, key)
		}
		if err != nil {
			return filter, err
		}
	}

	return filter, nil
}

// Validates uuid query param
//
// # Inputs:
//
//   - key [string] param name
//   - value [string] param value
//
// # Returns:
//
//   - The value
//   - ErrBadRequest when not a uuid
func uuidParam(key, value string) (string, error) {
	if !dto.IsValidUUID(value) {
		return "", fmt.Errorf("%w: %s %q is not a uuid", domain.ErrBadRequest, key, value)
	}
	return value, nil
}

// Splits comma separated list param
//
// # Inputs:
//
//   - key [string] param name
//   - value [string] comma separated values
//
// # Returns:
//
//   - List values
//   - ErrBadRequest when list empty
func listParam(key, value string) ([]string, error) {
	items := []string{}
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("%w: %s must list at least one value", domain.ErrBadRequest, key)
	}
	return items, nil
}

// Parses true/false query param
//
// # Inputs:
//
//   - key [string] param name
//   - value [string] param value
//
// # Returns:
//
//   - Parsed bool
//   - ErrBadRequest on anything but true/false
func boolParam(key, value string) (*bool, error) {
	switch strings.ToLower(value) {
	case "true":
		b := true
		return &b, nil
	case "false":
		b := false
		return &b, nil
	}
	return nil, fmt.Errorf("%w: %s must be true or false", domain.ErrBadRequest, key)
}

// Parses whole number query param
//
// # Inputs:
//
//   - key [string] param name
//   - value [string] param value
//
// # Returns:
//
//   - Parsed number
//   - ErrBadRequest when not a number
func intParam(key, value string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("%w: %s %q is not a number", domain.ErrBadRequest, key, value)
	}
	return n, nil
}
