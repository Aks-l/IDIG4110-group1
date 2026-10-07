package dto

import "regexp"

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsValidUUID reports whether s is a canonical uuid, the id format of every
// shared contract. Use it to reject malformed ids before they reach the
// database.
func IsValidUUID(s string) bool {
	return uuidPattern.MatchString(s)
}
