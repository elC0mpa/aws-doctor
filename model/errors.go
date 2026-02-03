// Package model defines the data structures used throughout the application.
package model

import "errors"

// ErrFirstDayOfMonth is returned when an operation cannot be performed on the first day of the month.
var ErrFirstDayOfMonth = errors.New("not available on the first day of the month")
