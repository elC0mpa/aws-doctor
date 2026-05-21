package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCacheService(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aws-doctor-cache-test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.RemoveAll(tempDir) }()

	s := &service{
		cacheDir: tempDir,
		ttls: map[Key]time.Duration{
			LatestVersionKey: time.Hour,
		},
	}

	key := LatestVersionKey
	val := "v1.0.0"

	var target string

	// Test Get empty
	found, err := s.Get(key, &target)
	if err != nil {
		t.Errorf("Get empty error: %v", err)
	}

	if found {
		t.Error("Expected found=false for missing key")
	}

	// Test Set
	err = s.Set(key, val)
	if err != nil {
		t.Errorf("Set error: %v", err)
	}

	// Test Get
	found, _ = s.Get(key, &target)

	if err != nil {
		t.Errorf("Get error: %v", err)
	}

	if !found || target != val {
		t.Errorf("Expected %s, got %s (found=%v)", val, target, found)
	}

	// Test expiration
	s.ttls[key] = -time.Hour

	found, err = s.Get(key, &target)
	if err != nil {
		t.Errorf("Get expired error: %v", err)
	}

	if found {
		t.Error("Expected found=false for expired key")
	}

	// Reset TTL for corruption test
	s.ttls[key] = time.Hour
	filePath := s.getFilePath(key)

	err = os.WriteFile(filePath, []byte("invalid-json"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Get(key, &target)
	if err == nil {
		t.Error("Expected error for corrupted JSON")
	}
}

func TestNewService_Sub(t *testing.T) {
	s := NewService()
	if s == nil {
		t.Error("NewService returned nil")
	}

	// Test Set/Get with contexts
	tempDir, _ := os.MkdirTemp("", "aws-doctor-cache-test-ctx")

	defer func() { _ = os.RemoveAll(tempDir) }()

	sc := &service{
		cacheDir: tempDir,
		ttls: map[Key]time.Duration{
			LatestVersionKey: time.Hour,
		},
	}

	val := "v2.0.0"

	var target string

	err := sc.Set(LatestVersionKey, val, "ctx1", "ctx2")
	if err != nil {
		t.Errorf("Set with context error: %v", err)
	}

	found, err := sc.Get(LatestVersionKey, &target, "ctx1", "ctx2")
	if err != nil || !found || target != val {
		t.Errorf("Get with context failed: found=%v, err=%v, target=%v", found, err, target)
	}
}

func TestCacheService_Errors(t *testing.T) {
	// Test Set error (read-only dir)
	if os.Getuid() == 0 {
		t.Skip("Skipping Set error test for root user")
	}

	s := &service{cacheDir: "/root/no-access"}

	err := s.Set(LatestVersionKey, "val")
	if err == nil {
		t.Error("Expected error for read-only directory")
	}
}

func TestCacheService_Corruption_Safe(t *testing.T) {
	tempDir, _ := os.MkdirTemp("", "aws-doctor-cache-test-corrupt-safe")

	defer func() { _ = os.RemoveAll(tempDir) }()

	s := &service{cacheDir: tempDir}

	filePath := s.getFilePath(LatestVersionKey)
	_ = os.MkdirAll(filepath.Dir(filePath), 0o755)
	_ = os.WriteFile(filePath, []byte("invalid-json"), 0o644)

	var target string

	found, _ := s.Get(LatestVersionKey, &target)
	if found {
		t.Error("Expected found=false for corrupted JSON")
	}
}

func TestCacheService_SetErrors_Safe(t *testing.T) {
	s := &service{cacheDir: "/root/no-access-aws-doctor"}

	err := s.Set(LatestVersionKey, "val")
	if err == nil {
		t.Error("Expected error for restricted directory")
	}
}
