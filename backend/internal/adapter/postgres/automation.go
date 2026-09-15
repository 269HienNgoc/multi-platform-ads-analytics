package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	applicationautomation "github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/application/automation"
	automationdomain "github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain/automation"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ applicationautomation.Store = (*AutomationStore)(nil)

type workflowRecord struct {
	ID                     string    `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID         string    `gorm:"column:organization_id"`
	AdAccountID            string    `gorm:"column:ad_account_id;type:uuid"`
	PageExternalID         string    `gorm:"column:page_external_id"`
	PixelExternalID        string    `gorm:"column:pixel_external_id"`
	PixelEvent             string    `gorm:"column:pixel_event"`
	ExistingPostID         string    `gorm:"column:existing_post_id"`
	SeedCampaignExternalID string    `gorm:"column:seed_campaign_external_id"`
	MainCampaignExternalID string    `gorm:"column:main_campaign_external_id"`
	SeedSpendLimitUSD      float64   `gorm:"column:seed_spend_limit_usd"`
	State                  string    `gorm:"column:state"`
	LastError              string    `gorm:"column:last_error"`
	CreatedAt              time.Time `gorm:"column:created_at"`
	UpdatedAt              time.Time `gorm:"column:updated_at"`
}

func (workflowRecord) TableName() string { return "campaign_workflows" }

type workflowMetricRecord struct {
	WorkflowID          string    `gorm:"column:workflow_id;type:uuid"`
	SpendUSD            float64   `gorm:"column:spend_usd"`
	Registrations       int64     `gorm:"column:registrations"`
	Deposits            int64     `gorm:"column:deposits"`
	CostPerRegistration float64   `gorm:"column:cost_per_registration"`
	CostPerDeposit      float64   `gorm:"column:cost_per_deposit"`
	CapturedAt          time.Time `gorm:"column:captured_at"`
	CreatedAt           time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (workflowMetricRecord) TableName() string { return "campaign_workflow_metrics" }

// AutomationStore persists campaign workflows and measurements with GORM.
type AutomationStore struct {
	db *gorm.DB
}

// NewAutomationStore creates a PostgreSQL campaign automation adapter.
func NewAutomationStore(client *Client) *AutomationStore {
	return &AutomationStore{db: client.db}
}

// List loads workflows in stable creation order.
func (s *AutomationStore) List(ctx context.Context) ([]automationdomain.CampaignWorkflow, error) {
	records := []workflowRecord{}
	if err := s.db.WithContext(ctx).Order("created_at, id").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("querying campaign workflows: %w", err)
	}

	workflows := make([]automationdomain.CampaignWorkflow, 0, len(records))
	for _, record := range records {
		workflows = append(workflows, workflowToDomain(record))
	}

	return workflows, nil
}

// CreateBulk persists all requested workflows atomically.
func (s *AutomationStore) CreateBulk(
	ctx context.Context,
	workflows []automationdomain.CampaignWorkflow,
) error {
	records := make([]workflowRecord, 0, len(workflows))
	for _, workflow := range workflows {
		records = append(records, workflowFromDomain(workflow))
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&records).Error; err != nil {
			return translateAutomationWriteError(err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// Update locks one workflow, applies domain logic, and persists the result atomically.
func (s *AutomationStore) Update(
	ctx context.Context,
	id string,
	change func(*automationdomain.CampaignWorkflow) error,
) (automationdomain.CampaignWorkflow, error) {
	return s.change(ctx, id, nil, change)
}

// RecordMetrics stores a measurement in the same transaction as any resulting state change.
func (s *AutomationStore) RecordMetrics(
	ctx context.Context,
	id string,
	metrics automationdomain.CampaignMetrics,
	change func(*automationdomain.CampaignWorkflow) error,
) (automationdomain.CampaignWorkflow, error) {
	return s.change(ctx, id, &metrics, change)
}

func (s *AutomationStore) change(
	ctx context.Context,
	id string,
	metrics *automationdomain.CampaignMetrics,
	change func(*automationdomain.CampaignWorkflow) error,
) (automationdomain.CampaignWorkflow, error) {
	var updated automationdomain.CampaignWorkflow
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var record workflowRecord
		query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).Take(&record)
		if query.Error != nil {
			return translateAutomationReadError(query.Error)
		}

		workflow := workflowToDomain(record)
		if err := change(&workflow); err != nil {
			return err
		}
		workflow.UpdatedAt = time.Now().UTC()
		if metrics != nil {
			metricRecord := metricFromDomain(id, *metrics)
			if err := tx.Create(&metricRecord).Error; err != nil {
				return fmt.Errorf("persisting campaign workflow metrics: %w", err)
			}
		}

		updates := map[string]any{
			"state":                     string(workflow.State),
			"seed_campaign_external_id": workflow.SeedCampaignExternalID,
			"main_campaign_external_id": workflow.MainCampaignExternalID,
			"last_error":                workflow.LastError,
			"updated_at":                workflow.UpdatedAt,
		}
		if err := tx.Model(&workflowRecord{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return fmt.Errorf("updating campaign workflow: %w", err)
		}
		updated = workflow

		return nil
	})
	if err != nil {
		return automationdomain.CampaignWorkflow{}, err
	}

	return updated, nil
}

func workflowFromDomain(workflow automationdomain.CampaignWorkflow) workflowRecord {
	return workflowRecord{
		ID: workflow.ID, OrganizationID: workflow.OrganizationID, AdAccountID: workflow.AdAccountID,
		PageExternalID: workflow.PageExternalID, PixelExternalID: workflow.PixelExternalID,
		PixelEvent: workflow.PixelEvent, ExistingPostID: workflow.ExistingPostID,
		SeedCampaignExternalID: workflow.SeedCampaignExternalID,
		MainCampaignExternalID: workflow.MainCampaignExternalID,
		SeedSpendLimitUSD:      workflow.SeedSpendLimitUSD, State: string(workflow.State),
		LastError: workflow.LastError, CreatedAt: workflow.CreatedAt, UpdatedAt: workflow.UpdatedAt,
	}
}

func workflowToDomain(record workflowRecord) automationdomain.CampaignWorkflow {
	return automationdomain.CampaignWorkflow{
		ID: record.ID, OrganizationID: record.OrganizationID, AdAccountID: record.AdAccountID,
		PageExternalID: record.PageExternalID, PixelExternalID: record.PixelExternalID,
		PixelEvent: record.PixelEvent, ExistingPostID: record.ExistingPostID,
		SeedCampaignExternalID: record.SeedCampaignExternalID,
		MainCampaignExternalID: record.MainCampaignExternalID,
		SeedSpendLimitUSD:      record.SeedSpendLimitUSD, State: automationdomain.WorkflowState(record.State),
		LastError: record.LastError, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
	}
}

func metricFromDomain(id string, metrics automationdomain.CampaignMetrics) workflowMetricRecord {
	return workflowMetricRecord{
		WorkflowID: id, SpendUSD: metrics.SpendUSD, Registrations: metrics.Registrations,
		Deposits: metrics.Deposits, CostPerRegistration: metrics.CostPerRegistration,
		CostPerDeposit: metrics.CostPerDeposit, CapturedAt: metrics.CapturedAt,
	}
}

func translateAutomationWriteError(err error) error {
	if errors.Is(err, gorm.ErrForeignKeyViolated) {
		return applicationautomation.ErrNotFound
	}

	return fmt.Errorf("persisting campaign workflows: %w", err)
}

func translateAutomationReadError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return applicationautomation.ErrNotFound
	}

	return fmt.Errorf("reading campaign workflow: %w", err)
}
