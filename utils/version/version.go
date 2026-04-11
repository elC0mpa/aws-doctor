package version

import "strings"

// IsEqual compares the latest tag from GitHub with the local version.
// It returns true if:
// 1. Both strings are exactly equal.
// 2. The tag equals the local version after removing its first character (e.g., 'v' prefix).
func IsEqual(latestTag, localVersion string) bool {
	if latestTag == localVersion {
		return true
	}

	if len(latestTag) > 0 && latestTag[1:] == localVersion {
		return true
	}

	return false
}

// HasVPrefix checks if the string starts with 'v'.
func HasVPrefix(s string) bool {
	return strings.HasPrefix(strings.ToLower(s), "v")
}
