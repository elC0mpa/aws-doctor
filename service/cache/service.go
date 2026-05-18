package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	versionCheckTTL = 2 * time.Hour
	appName         = "aws-doctor"
)

// NewService creates a new cache service.
func NewService() Service {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		cacheDir = os.TempDir()
	}

	return &service{
		cacheDir: filepath.Join(cacheDir, appName),
		ttls: map[Key]time.Duration{
			LatestVersionKey: versionCheckTTL,
		},
	}
}

func (s *service) Get(key Key, target interface{}, contexts ...string) (bool, error) {
	filePath := s.getFilePath(key, contexts...)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, fmt.Errorf("failed to read cache file: %w", err)
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return false, fmt.Errorf("failed to unmarshal cache data: %w", err)
	}

	ttl, ok := s.ttls[key]
	if !ok {
		return false, fmt.Errorf("no TTL defined for key: %s", key)
	}

	if time.Since(entry.Timestamp) > ttl {
		return false, nil
	}

	if err := json.Unmarshal(entry.Value, target); err != nil {
		return false, fmt.Errorf("failed to unmarshal value into target: %w", err)
	}

	return true, nil
}

func (s *service) Set(key Key, value interface{}, contexts ...string) error {
	if err := os.MkdirAll(s.cacheDir, 0o755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	valueJSON, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	entry := cacheEntry{
		Value:     valueJSON,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	filePath := s.getFilePath(key, contexts...)
	tmpPath := filePath + ".tmp"

	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write temp cache file: %w", err)
	}

	if err := os.Rename(tmpPath, filePath); err != nil {
		return fmt.Errorf("failed to commit cache file: %w", err)
	}

	return nil
}

func (s *service) getFilePath(key Key, contexts ...string) string {
	fileName := string(key)
	if len(contexts) > 0 {
		fileName = fmt.Sprintf("%s_%s", fileName, strings.Join(contexts, "_"))
	}

	return filepath.Join(s.cacheDir, fileName+".json")
}
