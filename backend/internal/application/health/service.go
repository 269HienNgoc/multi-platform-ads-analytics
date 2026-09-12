// Package health provides process and dependency health use cases.
package health

import "context"

const (
	// StatusUp indicates that a health check passed.
	StatusUp = "up"
	// StatusDown indicates that a health check failed.
	StatusDown = "down"
)

// Pinger is implemented by dependencies that can verify their readiness.
type Pinger interface {
	Ping(context.Context) error
}

// Result is the transport-independent health result.
type Result struct {
	Status string
	Checks map[string]string
}

// Service evaluates application liveness and readiness.
type Service struct {
	database Pinger
}

// New creates a health service.
func New(database Pinger) *Service {
	return &Service{database: database}
}

// Liveness reports whether the process is running.
func (s *Service) Liveness() Result {
	return Result{
		Status: StatusUp,
		Checks: map[string]string{"process": StatusUp},
	}
}

// Readiness verifies dependencies required to serve traffic.
func (s *Service) Readiness(ctx context.Context) (Result, error) {
	if err := s.database.Ping(ctx); err != nil {
		return Result{
			Status: StatusDown,
			Checks: map[string]string{"database": StatusDown},
		}, err
	}

	return Result{
		Status: StatusUp,
		Checks: map[string]string{"database": StatusUp},
	}, nil
}
