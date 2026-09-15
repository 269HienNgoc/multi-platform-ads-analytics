// Package automation coordinates durable campaign workflow use cases.
package automation

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	automationdomain "github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain/automation"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// ErrInvalidInput identifies an automation request that violates required invariants.
var ErrInvalidInput = errors.New("automation: invalid input")

// ErrNotFound identifies a requested workflow or advertising account that does not exist.
var ErrNotFound = errors.New("automation: not found")

// Store is the persistence contract consumed by campaign workflow use cases.
type Store interface {
	List(context.Context) ([]automationdomain.CampaignWorkflow, error)
	CreateBulk(context.Context, []automationdomain.CampaignWorkflow) error
	Update(
		context.Context,
		string,
		func(*automationdomain.CampaignWorkflow) error,
	) (automationdomain.CampaignWorkflow, error)
	RecordMetrics(
		context.Context,
		string,
		automationdomain.CampaignMetrics,
		func(*automationdomain.CampaignWorkflow) error,
	) (automationdomain.CampaignWorkflow, error)
}

// CreateBulkInput describes one workflow template applied independently to many accounts.
type CreateBulkInput struct {
	OrganizationID    string
	AdAccountIDs      []string
	PageExternalID    string
	PixelExternalID   string
	PixelEvent        string
	ExistingPostID    string
	SeedSpendLimitUSD float64
}

// Service implements provider-neutral campaign automation use cases.
type Service struct {
	store Store
	now   func() time.Time
}

// New creates a campaign automation service.
func New(store Store) *Service {
	return &Service{store: store, now: func() time.Time { return time.Now().UTC() }}
}

// List returns all workflows in stable creation order.
func (s *Service) List(ctx context.Context) ([]automationdomain.CampaignWorkflow, error) {
	workflows, err := s.store.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing workflows: %w", err)
	}
	if workflows == nil {
		return []automationdomain.CampaignWorkflow{}, nil
	}

	return workflows, nil
}

// CreateBulk creates one independent workflow for every selected account.
func (s *Service) CreateBulk(
	ctx context.Context,
	input CreateBulkInput,
) ([]automationdomain.CampaignWorkflow, error) {
	input.OrganizationID = strings.TrimSpace(input.OrganizationID)
	input.PageExternalID = strings.TrimSpace(input.PageExternalID)
	input.PixelExternalID = strings.TrimSpace(input.PixelExternalID)
	input.PixelEvent = strings.TrimSpace(input.PixelEvent)
	input.ExistingPostID = strings.TrimSpace(input.ExistingPostID)
	if input.OrganizationID == "" {
		return nil, invalidField("organization_id")
	}
	if input.PageExternalID == "" {
		return nil, invalidField("page_external_id")
	}
	if len(input.AdAccountIDs) == 0 {
		return nil, invalidField("ad_account_ids")
	}
	if input.SeedSpendLimitUSD < 0 {
		return nil, invalidField("seed_spend_limit_usd")
	}
	if input.SeedSpendLimitUSD == 0 {
		input.SeedSpendLimitUSD = 10
	}

	createdAt := s.now()
	seenAccounts := make(map[string]struct{}, len(input.AdAccountIDs))
	workflows := make([]automationdomain.CampaignWorkflow, 0, len(input.AdAccountIDs))
	for _, rawAccountID := range input.AdAccountIDs {
		accountID := strings.ToLower(strings.TrimSpace(rawAccountID))
		if !uuidPattern.MatchString(accountID) {
			return nil, invalidField("ad_account_ids")
		}
		if _, exists := seenAccounts[accountID]; exists {
			return nil, invalidField("ad_account_ids")
		}
		seenAccounts[accountID] = struct{}{}

		id, err := newUUID()
		if err != nil {
			return nil, fmt.Errorf("generating workflow id: %w", err)
		}
		workflows = append(workflows, automationdomain.CampaignWorkflow{
			ID: id, OrganizationID: input.OrganizationID, AdAccountID: accountID,
			PageExternalID: input.PageExternalID, PixelExternalID: input.PixelExternalID,
			PixelEvent: input.PixelEvent, ExistingPostID: input.ExistingPostID,
			SeedSpendLimitUSD: input.SeedSpendLimitUSD, State: automationdomain.StateSeedPending,
			CreatedAt: createdAt, UpdatedAt: createdAt,
		})
	}

	if err := s.store.CreateBulk(ctx, workflows); err != nil {
		return nil, fmt.Errorf("creating workflows: %w", err)
	}

	return workflows, nil
}

// Transition applies one valid explicit state transition.
func (s *Service) Transition(
	ctx context.Context,
	id string,
	to automationdomain.WorkflowState,
) (automationdomain.CampaignWorkflow, error) {
	id = strings.ToLower(strings.TrimSpace(id))
	if !uuidPattern.MatchString(id) || !to.IsValid() {
		return automationdomain.CampaignWorkflow{}, invalidField("workflow transition")
	}

	workflow, err := s.store.Update(ctx, id, func(item *automationdomain.CampaignWorkflow) error {
		if err := automationdomain.ValidateTransition(item.State, to); err != nil {
			return err
		}
		item.State = to

		return nil
	})
	if err != nil {
		return automationdomain.CampaignWorkflow{}, fmt.Errorf("transitioning workflow: %w", err)
	}

	return workflow, nil
}

// ApplyMetrics records a measurement and advances a completed seed workflow.
func (s *Service) ApplyMetrics(
	ctx context.Context,
	id string,
	metrics automationdomain.CampaignMetrics,
) (automationdomain.CampaignWorkflow, error) {
	id = strings.ToLower(strings.TrimSpace(id))
	if !uuidPattern.MatchString(id) || metrics.SpendUSD < 0 || metrics.Registrations < 0 || metrics.Deposits < 0 ||
		metrics.CostPerRegistration < 0 || metrics.CostPerDeposit < 0 {
		return automationdomain.CampaignWorkflow{}, invalidField("metrics")
	}
	if metrics.CapturedAt.IsZero() {
		metrics.CapturedAt = s.now()
	}

	workflow, err := s.store.RecordMetrics(
		ctx,
		id,
		metrics,
		func(item *automationdomain.CampaignWorkflow) error {
			isSeedComplete := item.State == automationdomain.StateSeedRunning &&
				metrics.SpendUSD >= item.SeedSpendLimitUSD
			if !isSeedComplete {
				return nil
			}
			if err := automationdomain.ValidateTransition(item.State, automationdomain.StateSeedCompleted); err != nil {
				return err
			}
			if err := automationdomain.ValidateTransition(
				automationdomain.StateSeedCompleted,
				automationdomain.StateMainPending,
			); err != nil {
				return err
			}
			item.State = automationdomain.StateMainPending

			return nil
		},
	)
	if err != nil {
		return automationdomain.CampaignWorkflow{}, fmt.Errorf("applying workflow metrics: %w", err)
	}

	return workflow, nil
}

func invalidField(field string) error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, field)
}

func newUUID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("reading secure random bytes: %w", err)
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80

	return fmt.Sprintf(
		"%x-%x-%x-%x-%x",
		value[0:4], value[4:6], value[6:8], value[8:10], value[10:16],
	), nil
}
