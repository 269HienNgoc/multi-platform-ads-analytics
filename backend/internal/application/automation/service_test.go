package automation

import (
	"context"
	"errors"
	"testing"
	"time"

	automationdomain "github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain/automation"
)

type storeStub struct {
	workflows      []automationdomain.CampaignWorkflow
	missingAccount string
	err            error
}

func (s *storeStub) List(context.Context) ([]automationdomain.CampaignWorkflow, error) {
	return append([]automationdomain.CampaignWorkflow{}, s.workflows...), s.err
}

func (s *storeStub) Create(
	_ context.Context,
	workflow automationdomain.CampaignWorkflow,
) (automationdomain.CampaignWorkflow, error) {
	if s.err != nil {
		return automationdomain.CampaignWorkflow{}, s.err
	}
	if workflow.AdAccountID == s.missingAccount {
		return automationdomain.CampaignWorkflow{}, ErrNotFound
	}
	s.workflows = append(s.workflows, workflow)

	return workflow, nil
}

func TestServiceCreateBulkKeepsSuccessfulAccounts(t *testing.T) {
	t.Parallel()

	existingAccount := "00000000-0000-4000-8000-000000000001"
	missingAccount := "00000000-0000-4000-8000-000000000002"
	store := &storeStub{missingAccount: missingAccount}
	service := New(store)
	result, err := service.CreateBulk(t.Context(), CreateBulkInput{
		RequestKey: "partial-request", OrganizationID: "org",
		AdAccountIDs: []string{existingAccount, missingAccount}, PageExternalID: "page",
	})
	if err != nil {
		t.Fatalf("CreateBulk() error = %v", err)
	}
	if len(result.Workflows) != 1 || result.Workflows[0].AdAccountID != existingAccount {
		t.Errorf("CreateBulk() workflows = %#v, expected successful account", result.Workflows)
	}
	if len(result.Failures) != 1 || result.Failures[0].AdAccountID != missingAccount {
		t.Errorf("CreateBulk() failures = %#v, expected missing account", result.Failures)
	}
}

func (s *storeStub) Update(
	_ context.Context,
	id string,
	change func(*automationdomain.CampaignWorkflow) error,
) (automationdomain.CampaignWorkflow, error) {
	return s.change(id, change)
}

func (s *storeStub) RecordMetrics(
	_ context.Context,
	id string,
	_ automationdomain.CampaignMetrics,
	change func(*automationdomain.CampaignWorkflow) error,
) (automationdomain.CampaignWorkflow, error) {
	return s.change(id, change)
}

func (s *storeStub) change(
	id string,
	change func(*automationdomain.CampaignWorkflow) error,
) (automationdomain.CampaignWorkflow, error) {
	for index := range s.workflows {
		if s.workflows[index].ID != id {
			continue
		}
		if err := change(&s.workflows[index]); err != nil {
			return automationdomain.CampaignWorkflow{}, err
		}

		return s.workflows[index], nil
	}

	return automationdomain.CampaignWorkflow{}, ErrNotFound
}

func TestService_CreateBulk(t *testing.T) {
	t.Parallel()

	accountID := "00000000-0000-4000-8000-000000000001"
	tests := []struct {
		name        string
		input       CreateBulkInput
		expectedErr error
	}{
		{
			name: "creates one workflow per account",
			input: CreateBulkInput{
				RequestKey: "request-1", OrganizationID: "org",
				AdAccountIDs: []string{accountID}, PageExternalID: "page",
			},
		},
		{name: "requires organization", input: CreateBulkInput{}, expectedErr: ErrInvalidInput},
		{
			name: "rejects duplicate accounts",
			input: CreateBulkInput{
				RequestKey: "request-1", OrganizationID: "org",
				AdAccountIDs: []string{accountID, accountID}, PageExternalID: "page",
			},
			expectedErr: ErrInvalidInput,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store := &storeStub{}
			service := New(store)
			service.now = func() time.Time { return time.Unix(1, 0).UTC() }
			created, err := service.CreateBulk(t.Context(), test.input)
			if !errors.Is(err, test.expectedErr) {
				t.Errorf("CreateBulk() error = %v, expected %v", err, test.expectedErr)
			}
			if test.expectedErr == nil &&
				(len(created.Workflows) != 1 || created.Workflows[0].SeedSpendLimitUSD != 10) {
				t.Errorf("CreateBulk() = %#v, expected one workflow with default limit", created)
			}
		})
	}
}

func TestService_TransitionAndApplyMetrics(t *testing.T) {
	t.Parallel()

	workflowID := "00000000-0000-4000-8000-000000000002"
	store := &storeStub{workflows: []automationdomain.CampaignWorkflow{{
		ID: workflowID, State: automationdomain.StateSeedPending, SeedSpendLimitUSD: 10,
	}}}
	service := New(store)

	workflow, err := service.Transition(t.Context(), workflowID, automationdomain.StateSeedCreating)
	if err != nil {
		t.Fatalf("Transition() error = %v", err)
	}
	if workflow.State != automationdomain.StateSeedCreating {
		t.Errorf("state = %q, expected %q", workflow.State, automationdomain.StateSeedCreating)
	}
	if _, err := service.Transition(t.Context(), workflowID, automationdomain.StateSeedRunning); err != nil {
		t.Fatalf("Transition() to running error = %v", err)
	}

	workflow, err = service.ApplyMetrics(t.Context(), workflowID, automationdomain.CampaignMetrics{SpendUSD: 10})
	if err != nil {
		t.Fatalf("ApplyMetrics() error = %v", err)
	}
	if workflow.State != automationdomain.StateMainPending {
		t.Errorf("state = %q, expected %q", workflow.State, automationdomain.StateMainPending)
	}
}
