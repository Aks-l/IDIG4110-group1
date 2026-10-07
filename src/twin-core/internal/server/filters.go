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

// parseEntityFilter validates the /state list endpoint's query params.
// Unknown keys, repeated params, and malformed values fail with
// ErrBadRequest: typos should fail loudly instead of silently dropping a
// filter.
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

// uuidParam validates a uuid query param.
func uuidParam(key, value string) (string, error) {
	if !dto.IsValidUUID(value) {
		return "", fmt.Errorf("%w: %s %q is not a uuid", domain.ErrBadRequest, key, value)
	}
	return value, nil
}

// listParam splits a comma-separated list param, e.g. "sensor,switch".
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

// boolParam accepts only true / false.
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

// intParam parses a whole number.
func intParam(key, value string) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("%w: %s %q is not a number", domain.ErrBadRequest, key, value)
	}
	return n, nil
}
