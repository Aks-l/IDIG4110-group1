package domain

import (
	"strings"
	"unicode/utf8"
)

// EntityDomain derives an entity's domain from the source's entity id the
// way Home Assistant shapes its ids ("sensor.living_room_temperature" ->
// "sensor").
func EntityDomain(externalEntityID string) string {
	domain := externalEntityID
	if i := strings.IndexByte(domain, '.'); i > 0 {
		domain = domain[:i]
	}
	if domain == "" {
		domain = "unknown"
	}
	return TruncateRunes(domain, 50)
}

// TruncateRunes limits s to max characters, staying on rune boundaries.
func TruncateRunes(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max])
}
