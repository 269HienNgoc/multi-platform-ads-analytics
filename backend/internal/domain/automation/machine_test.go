package automation

import (
	"errors"
	"testing"
)

func TestValidateTransition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		from      WorkflowState
		to        WorkflowState
		isInvalid bool
	}{
		{name: "advance seed", from: StateSeedPending, to: StateSeedCreating},
		{name: "same state", from: StateMainRunning, to: StateMainRunning},
		{name: "skip seed", from: StateSeedPending, to: StateMainRunning, isInvalid: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateTransition(test.from, test.to)
			if errors.Is(err, ErrInvalidTransition) != test.isInvalid {
				t.Errorf("ValidateTransition() error = %v, invalid = %t", err, test.isInvalid)
			}
		})
	}
}
