package rule

import (
	"strconv"
	"strings"

	"IDIG4110/shared/dto"
)

// Refers reports whether m is about the entity of reading r.
func (m Match) Refers(r dto.Reading) bool {
	return m.Entity == r.ExternalEntityID && (m.GatewayID == "" || m.GatewayID == r.GatewayID)
}

// Matches compares the value of reading r with m. Numbers are compared
// numerically, also when the reading carries the number as text. A boolean
// rule value stands for the "on"/"off" states that binary sensors report.
func (m Match) Matches(r dto.Reading) bool {
	switch want := m.Value.(type) {
	case float64:
		got, ok := numericValue(r)
		if !ok {
			return false
		}
		return compareNumbers(m.Operator, got, want)
	case bool:
		state := "off"
		if want {
			state = "on"
		}
		return compareText(m.Operator, textValue(r), state)
	case string:
		return compareText(m.Operator, textValue(r), want)
	}
	return false
}

func numericValue(r dto.Reading) (float64, bool) {
	if r.ValueNum != nil {
		return *r.ValueNum, true
	}
	if r.ValueText != nil {
		f, err := strconv.ParseFloat(strings.TrimSpace(*r.ValueText), 64)
		return f, err == nil
	}
	return 0, false
}

func textValue(r dto.Reading) string {
	if r.ValueText != nil {
		return *r.ValueText
	}
	if r.ValueNum != nil {
		return strconv.FormatFloat(*r.ValueNum, 'f', -1, 64)
	}
	return ""
}

func compareNumbers(op Operator, got, want float64) bool {
	switch op {
	case OpEq:
		return got == want
	case OpNe:
		return got != want
	case OpGt:
		return got > want
	case OpGte:
		return got >= want
	case OpLt:
		return got < want
	case OpLte:
		return got <= want
	}
	return false
}

func compareText(op Operator, got, want string) bool {
	equal := strings.EqualFold(strings.TrimSpace(got), want)
	switch op {
	case OpEq:
		return equal
	case OpNe:
		return !equal
	}
	return false
}
