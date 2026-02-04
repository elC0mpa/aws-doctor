package update

// Service is the interface for the update service.
type Service interface {
	Update() error
}

type commandRunner interface {
	Run(name string, arg ...string) error
}

type service struct {
	runner commandRunner
}
