package flag

import "github.com/elC0mpa/aws-doctor/model"

type service struct{}

// Service is the interface for CLI flag service.
type Service interface {
	GetParsedFlags() (model.Flags, error)
}
