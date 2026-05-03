package model

import "time"

// UnusedSecretInfo represents information about a secret that has not been accessed recently.
type UnusedSecretInfo struct {
	Name             string
	LastAccessedDate *time.Time
}
