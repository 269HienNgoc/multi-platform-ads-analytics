package workflow

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain"
)

var ErrNotFound = errors.New("workflow not found")

type Service struct {
	mu        sync.RWMutex
	workflows map[string]*domain.CampaignWorkflow
}

func NewService() *Service {
	return &Service{workflows: make(map[string]*domain.CampaignWorkflow)}
}

func (s *Service) List() []domain.CampaignWorkflow {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]domain.CampaignWorkflow, 0, len(s.workflows))
	for _, item := range s.workflows {
		result = append(result, *item)
	}
	return result
}

func (s *Service) CreateBulk(req domain.BulkWorkflowRequest) ([]domain.CampaignWorkflow, error) {
	if req.OrganizationID == "" || req.PageID == "" || len(req.AdAccountIDs) == 0 {
		return nil, errors.New("organizationId, pageId and adAccountIds are required")
	}
	if req.SeedSpendLimitUSD <= 0 {
		req.SeedSpendLimitUSD = 10
	}

	now := time.Now().UTC()
	created := make([]domain.CampaignWorkflow, 0, len(req.AdAccountIDs))

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, accountID := range req.AdAccountIDs {
		item := &domain.CampaignWorkflow{
			ID: randomID("wf"),
			OrganizationID: req.OrganizationID,
			AdAccountID: accountID,
			PageID: req.PageID,
			PixelID: req.PixelID,
			PixelEvent: req.PixelEvent,
			ExistingPostID: req.ExistingPostID,
			SeedSpendLimitUSD: req.SeedSpendLimitUSD,
			State: domain.StateSeedPending,
			CreatedAt: now,
			UpdatedAt: now,
		}
		s.workflows[item.ID] = item
		created = append(created, *item)
	}
	return created, nil
}

func (s *Service) Transition(id string, to domain.WorkflowState) (domain.CampaignWorkflow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.workflows[id]
	if !ok {
		return domain.CampaignWorkflow{}, ErrNotFound
	}
	if err := Transition(item.State, to); err != nil {
		return domain.CampaignWorkflow{}, err
	}
	item.State = to
	item.UpdatedAt = time.Now().UTC()
	return *item, nil
}

func (s *Service) ApplyMetrics(id string, metrics domain.CampaignMetrics) (domain.CampaignWorkflow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.workflows[id]
	if !ok {
		return domain.CampaignWorkflow{}, ErrNotFound
	}

	if item.State == domain.StateSeedRunning && metrics.SpendUSD >= item.SeedSpendLimitUSD {
		item.State = domain.StateSeedCompleted
		item.UpdatedAt = time.Now().UTC()
		item.State = domain.StateMainPending
		item.UpdatedAt = time.Now().UTC()
	}
	return *item, nil
}

func randomID(prefix string) string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return prefix + "-fallback"
	}
	return prefix + "-" + hex.EncodeToString(buf)
}
