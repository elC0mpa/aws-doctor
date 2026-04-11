package version

import "strings"

// IsEqual compares the latest tag from GitHub with the local version.
func IsEqual(latestTag, localVersion string) bool {
	l := strings.TrimPrefix(strings.ToLower(latestTag), "v")
	v := strings.TrimPrefix(strings.ToLower(localVersion), "v")

	return l == v
}
