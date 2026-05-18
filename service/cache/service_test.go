package cache

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheService(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aws-doctor-cache-test")
	assert.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	s := &service{
		cacheDir: tempDir,
		ttls: map[Key]time.Duration{
			"test_key": 1 * time.Second,
		},
	}

	key := Key("test_key")
	value := "test_value"

	// Test Set
	err = s.Set(key, value)
	assert.NoError(t, err)

	// Test Get (Success)
	var got string

	found, err := s.Get(key, &got)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, value, got)

	// Test Get (Non-existent)
	var gotNone string

	found, err = s.Get("none", &gotNone)
	assert.NoError(t, err)
	assert.False(t, found)

	// Test Expiration
	time.Sleep(1100 * time.Millisecond)

	var gotExpired string

	found, err = s.Get(key, &gotExpired)
	assert.NoError(t, err)
	assert.False(t, found)
}

func TestCacheService_WithContext(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aws-doctor-cache-test-ctx")
	assert.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	s := &service{
		cacheDir: tempDir,
		ttls: map[Key]time.Duration{
			"ctx_key": 1 * time.Hour,
		},
	}

	key := Key("ctx_key")
	value := "ctx_value"
	ctx := "us-east-1"

	// Test Set with context
	err = s.Set(key, value, ctx)
	assert.NoError(t, err)

	// Test Get with same context (Success)
	var got string

	found, err := s.Get(key, &got, ctx)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, value, got)

	// Test Get with different context (Fail)
	var gotOther string

	found, err = s.Get(key, &gotOther, "us-west-2")
	assert.NoError(t, err)
	assert.False(t, found)
}

func TestCacheService_ComplexObject(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aws-doctor-cache-test-complex")
	assert.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	s := &service{
		cacheDir: tempDir,
		ttls: map[Key]time.Duration{
			"complex_key": 1 * time.Hour,
		},
	}

	type complex struct {
		ID   int
		Name string
	}

	key := Key("complex_key")
	value := complex{ID: 1, Name: "test"}

	err = s.Set(key, value)
	assert.NoError(t, err)

	var got complex

	found, err := s.Get(key, &got)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, value, got)
}
