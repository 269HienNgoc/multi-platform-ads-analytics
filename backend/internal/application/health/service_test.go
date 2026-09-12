package health

import (
	"context"
	"errors"
	"testing"
)

var errDatabaseUnavailable = errors.New("database unavailable")

type pingerStub struct {
	err error
}

func (p pingerStub) Ping(context.Context) error {
	return p.err
}

func TestServiceLiveness(t *testing.T) {
	t.Parallel()

	service := New(pingerStub{})
	result := service.Liveness()

	if result.Status != StatusUp || result.Checks["process"] != StatusUp {
		t.Errorf("Liveness() = %#v, expected process up", result)
	}
}

func TestServiceReadiness(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		pingErr        error
		expectedStatus string
		expectedErr    error
	}{
		{name: "database ready", expectedStatus: StatusUp},
		{name: "database unavailable", pingErr: errDatabaseUnavailable, expectedStatus: StatusDown, expectedErr: errDatabaseUnavailable},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := New(pingerStub{err: test.pingErr})
			result, err := service.Readiness(t.Context())
			if !errors.Is(err, test.expectedErr) {
				t.Fatalf("Readiness() error = %v, expected %v", err, test.expectedErr)
			}
			if result.Status != test.expectedStatus {
				t.Errorf("Readiness().Status = %q, expected %q", result.Status, test.expectedStatus)
			}
		})
	}
}
