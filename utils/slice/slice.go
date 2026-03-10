package slice

import (
	"strings"
)

// ContainsIgnoreCase checks if a string slice contains a string, ignoring case and leading/trailing whitespace.
func ContainsIgnoreCase(slice []string, val string) bool {
	for _, item := range slice {
		if strings.EqualFold(strings.TrimSpace(item), val) {
			return true
		}
	}
	return false
}
