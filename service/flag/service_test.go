package flag

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetParsedFlags(t *testing.T) {
	// Reset the global flag set
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Set dummy command line arguments
	os.Args = []string{"cmd", "-update", "-region", "us-east-1"}

	svc := NewService()
	flags, err := svc.GetParsedFlags()

	assert.NoError(t, err)
	assert.True(t, flags.Update)
	assert.Equal(t, "us-east-1", flags.Region)
	assert.False(t, flags.Trend)
	assert.False(t, flags.Waste)
	assert.False(t, flags.Version)
}
