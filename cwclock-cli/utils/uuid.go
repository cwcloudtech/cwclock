package utils

import (
	"regexp"
	"strings"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsValidUUID reports whether value is shaped like a UUID (8-4-4-4-12 hex
// digits). Shared by the CLI's id-or-name resolution (a real id skips the
// extra list+match API call a name lookup needs) and anywhere a value
// needs to be a genuine id before being sent as one (e.g. a "clientId"
// filter query param).
func IsValidUUID(value string) bool {
	return uuidPattern.MatchString(strings.TrimSpace(value))
}
