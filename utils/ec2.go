package utils //nolint:revive

import (
	"fmt"
	"regexp"
	"time"
)

var transitionReasonRegex = regexp.MustCompile(`\(([^)]+)\)`)

// ParseTransitionDate parses the date from the state transition reason string.
func ParseTransitionDate(reason string) (time.Time, error) {
	matches := transitionReasonRegex.FindStringSubmatch(reason)
	if len(matches) < 2 {
		return time.Time{}, fmt.Errorf("no date found in string: %s", reason)
	}

	dateStr := matches[1]

	layout := "2006-01-02 15:04:05 MST"
	return time.Parse(layout, dateStr)
}
