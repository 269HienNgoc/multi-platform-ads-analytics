package workflow

import (
	"testing"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain"
)

func TestValidSeedFlow(t *testing.T) {
	states := []domain.WorkflowState{
		domain.StateSeedPending,
		domain.StateSeedCreating,
		domain.StateSeedRunning,
		domain.StateSeedCompleted,
		domain.StateMainPending,
		domain.StateMainValidating,
		domain.StateMainCreating,
		domain.StateMainRunning,
	}

	for i := 0; i < len(states)-1; i++ {
		if err := Transition(states[i], states[i+1]); err != nil {
			t.Fatalf("expected valid transition: %v", err)
		}
	}
}

func TestRejectsSkippingSeedState(t *testing.T) {
	if err := Transition(domain.StateSeedPending, domain.StateMainRunning); err == nil {
		t.Fatal("expected invalid transition error")
	}
}
