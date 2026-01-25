package utils //nolint:revive

import (
	"strconv"
	"strings"
)

// ParseCostString parses a cost string like "123.45 USD" into a float64
func ParseCostString(costStr string) float64 {
	parts := strings.Split(costStr, " ")
	if len(parts) == 0 {
		return 0
	}

	amount, _ := strconv.ParseFloat(parts[0], 64)

	return amount
}
