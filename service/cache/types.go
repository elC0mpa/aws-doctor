package cache

import (
	"encoding/json"
	"time"
)

// Key represents a specific feature's cache key.
type Key string

const (
	// LatestVersionKey is the key for the GitHub version check.
	LatestVersionKey Key = "latest_version"
)

// Service is the interface for the cache service.
type Service interface {
	Get(key Key, target interface{}, contexts ...string) (bool, error)
	Set(key Key, value interface{}, contexts ...string) error
}

type cacheEntry struct {
	Value     json.RawMessage `json:"value"`
	Timestamp time.Time       `json:"timestamp"`
}

type service struct {
	cacheDir string
	ttls     map[Key]time.Duration
}
