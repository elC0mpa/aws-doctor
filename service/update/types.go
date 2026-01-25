package update

// Service is the interface for the update service.
type Service interface {
	Update() error
}

type service struct{}
