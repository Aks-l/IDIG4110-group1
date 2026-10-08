package dto

import "regexp"

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Reports whether s is a canonical uuid
//
// # Inputs:
//
//   - s [string] value to check
//
// # Returns:
//
//   - True when s is a uuid
func IsValidUUID(s string) bool {
	return uuidPattern.MatchString(s)
}
