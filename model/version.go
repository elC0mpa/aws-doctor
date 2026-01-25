// Package model defines the data structures used throughout the application.
package model

// VersionInfo contains build-time metadata about the application.
type VersionInfo struct {
	Version string
	Commit  string
	Date    string
}
