package flag

import "github.com/elC0mpa/aws-doctor/model"

type service struct{}

type FlagService interface {
	GetParsedFlags() (model.Flags, error)
}
