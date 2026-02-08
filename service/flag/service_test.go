package flag

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetParsedFlags(t *testing.T) {
	svc := NewService()
	flags, err := svc.GetParsedFlags([]string{"-update", "-region", "us-east-1"})

	assert.NoError(t, err)
	assert.True(t, flags.Update)
	assert.Equal(t, "us-east-1", flags.Region)
	assert.False(t, flags.Trend)
	assert.False(t, flags.Waste)
	assert.False(t, flags.Version)
}
