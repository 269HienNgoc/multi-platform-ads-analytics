package workflow

import (
	"fmt"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain"
)

var allowedTransitions = map[domain.WorkflowState]map[domain.WorkflowState]bool{
	domain.StateAccountConnected: {domain.StateAssetsSynced: true, domain.StateFailed: true},
	domain.StateAssetsSynced: {domain.StatePreflightPassed: true, domain.StateManualReview: true, domain.StateFailed: true},
	domain.StatePreflightPassed: {domain.StateSeedPending: true, domain.StateMainPending: true, domain.StateFailed: true},
	domain.StateSeedPending: {domain.StateSeedCreating: true, domain.StateFailed: true},
	domain.StateSeedCreating: {domain.StateSeedRunning: true, domain.StateManualReview: true, domain.StateFailed: true},
	domain.StateSeedRunning: {domain.StateSeedCompleted: true, domain.StatePaused: true, domain.StateManualReview: true, domain.StateFailed: true},
	domain.StateSeedCompleted: {domain.StateMainPending: true, domain.StateFailed: true},
	domain.StateMainPending: {domain.StateMainValidating: true, domain.StateFailed: true},
	domain.StateMainValidating: {domain.StateMainCreating: true, domain.StateManualReview: true, domain.StateFailed: true},
	domain.StateMainCreating: {domain.StateMainRunning: true, domain.StateManualReview: true, domain.StateFailed: true},
	domain.StateMainRunning: {domain.StatePaused: true, domain.StateManualReview: true, domain.StateFailed: true},
	domain.StatePaused: {domain.StateMainRunning: true, domain.StateManualReview: true},
	domain.StateManualReview: {domain.StateSeedPending: true, domain.StateMainPending: true, domain.StatePaused: true, domain.StateFailed: true},
}

func Transition(from, to domain.WorkflowState) error {
	if from == to {
		return nil
	}
	if !allowedTransitions[from][to] {
		return fmt.Errorf("invalid workflow transition: %s -> %s", from, to)
	}
	return nil
}
