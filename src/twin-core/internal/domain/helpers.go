package domain

import (
	"strings"
	"unicode/utf8"
)

// Derives entity domain from HA shaped id ("sensor.temp" -> "sensor")
//
// # Inputs:
//
//   - externalEntityID [string] source entity id
//
// # Returns:
//
//   - Domain part, "unknown" when missing, max 50 chars
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

// Limits s to max characters on rune boundaries
//
// # Inputs:
//
//   - s [string] value to truncate
//   - max [int] max rune count
//
// # Returns:
//
//   - Truncated string
func TruncateRunes(s string, max int) string {
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max])
}
