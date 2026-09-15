package automation

import (
	"errors"
	"fmt"
)

// ErrInvalidTransition identifies a state change that violates the workflow lifecycle.
var ErrInvalidTransition = errors.New("automation: invalid workflow transition")

var allowedTransitions = map[WorkflowState]map[WorkflowState]struct{}{
	StateAccountConnected: {StateAssetsSynced: {}, StateFailed: {}},
	StateAssetsSynced:     {StatePreflightPassed: {}, StateManualReview: {}, StateFailed: {}},
	StatePreflightPassed:  {StateSeedPending: {}, StateMainPending: {}, StateFailed: {}},
	StateSeedPending:      {StateSeedCreating: {}, StateFailed: {}},
	StateSeedCreating:     {StateSeedRunning: {}, StateManualReview: {}, StateFailed: {}},
	StateSeedRunning:      {StateSeedCompleted: {}, StatePaused: {}, StateManualReview: {}, StateFailed: {}},
	StateSeedCompleted:    {StateMainPending: {}, StateFailed: {}},
	StateMainPending:      {StateMainValidating: {}, StateFailed: {}},
	StateMainValidating:   {StateMainCreating: {}, StateManualReview: {}, StateFailed: {}},
	StateMainCreating:     {StateMainRunning: {}, StateManualReview: {}, StateFailed: {}},
	StateMainRunning:      {StatePaused: {}, StateManualReview: {}, StateFailed: {}},
	StatePaused:           {StateMainRunning: {}, StateManualReview: {}},
	StateManualReview:     {StateSeedPending: {}, StateMainPending: {}, StatePaused: {}, StateFailed: {}},
}

// ValidateTransition verifies a requested workflow state change.
func ValidateTransition(from, to WorkflowState) error {
	if from == to {
		return nil
	}
	if _, allowed := allowedTransitions[from][to]; !allowed {
		return fmt.Errorf("%w: %s to %s", ErrInvalidTransition, from, to)
	}

	return nil
}
