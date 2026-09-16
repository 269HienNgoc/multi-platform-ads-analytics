package postgres

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/application/connector"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/application/providersync"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain/ads"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ providersync.Store = (*ProviderSyncStore)(nil)

type syncRunRecord struct {
	ID               string     `gorm:"column:id;type:uuid;primaryKey"`
	PlatformCode     string     `gorm:"column:platform_code"`
	Status           string     `gorm:"column:status"`
	CursorValue      string     `gorm:"column:cursor_value"`
	StartedAt        *time.Time `gorm:"column:started_at"`
	FinishedAt       *time.Time `gorm:"column:finished_at"`
	RecordsProcessed int64      `gorm:"column:records_processed"`
	ErrorSummary     string     `gorm:"column:error_summary"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (syncRunRecord) TableName() string { return "sync_runs" }

type rawPayloadRecord struct {
	ID          string          `gorm:"column:id;type:uuid;primaryKey"`
	SyncRunID   string          `gorm:"column:sync_run_id;type:uuid"`
	EntityKind  string          `gorm:"column:entity_kind"`
	ExternalID  string          `gorm:"column:external_id"`
	Payload     json.RawMessage `gorm:"column:payload;type:jsonb"`
	PayloadHash string          `gorm:"column:payload_hash"`
	ReceivedAt  time.Time       `gorm:"column:received_at"`
}

func (rawPayloadRecord) TableName() string { return "raw_provider_payloads" }

// ProviderSyncStore persists provider sync runs and normalized catalog snapshots.
type ProviderSyncStore struct {
	db *gorm.DB
}

// NewProviderSyncStore creates a PostgreSQL provider synchronization adapter.
func NewProviderSyncStore(client *Client) *ProviderSyncStore {
	return &ProviderSyncStore{db: client.db}
}

// CreateRun creates the observable running state before calling the provider.
func (s *ProviderSyncStore) CreateRun(ctx context.Context, run providersync.Run) error {
	record := syncRunRecord{
		ID: run.ID, PlatformCode: string(run.Platform), Status: run.Status,
		StartedAt: run.StartedAt, CreatedAt: run.CreatedAt, UpdatedAt: run.UpdatedAt,
	}
	if err := s.db.WithContext(ctx).Create(&record).Error; err != nil {
		return fmt.Errorf("persisting sync run: %w", err)
	}

	return nil
}

// ApplySnapshot upserts all accounts and campaigns in one catalog transaction.
func (s *ProviderSyncStore) ApplySnapshot(
	ctx context.Context,
	runID string,
	platform ads.Platform,
	snapshot connector.AssetSnapshot,
	syncedAt time.Time,
) (providersync.ApplyResult, error) {
	result := providersync.ApplyResult{
		Accounts: len(snapshot.AdAccounts), Campaigns: len(snapshot.Campaigns),
		Pages: len(snapshot.Pages), Pixels: len(snapshot.Pixels),
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		accountIDs := make(map[string]string, len(snapshot.AdAccounts))
		for _, account := range snapshot.AdAccounts {
			accountID, err := upsertProviderAccount(tx, platform, account, syncedAt)
			if err != nil {
				return err
			}
			accountIDs[account.ExternalID] = accountID
			if err := storeRawPayload(tx, runID, "account", account.ExternalID, account.Raw, syncedAt); err != nil {
				return err
			}
		}

		for _, campaign := range snapshot.Campaigns {
			accountID, exists := accountIDs[campaign.AccountExternalID]
			if !exists {
				return fmt.Errorf("campaign %s references an unknown provider account", campaign.ExternalID)
			}
			if err := upsertProviderCampaign(tx, accountID, campaign, syncedAt); err != nil {
				return err
			}
			if err := storeRawPayload(tx, runID, "campaign", campaign.ExternalID, campaign.Raw, syncedAt); err != nil {
				return err
			}
		}
		for _, accountID := range accountIDs {
			if err := tx.Model(&campaignRecord{}).
				Where(
					"account_id = ? AND last_synced_at IS NOT NULL AND last_synced_at <> ?",
					accountID,
					syncedAt,
				).
				Updates(map[string]any{"status": string(ads.StatusArchived), "updated_at": syncedAt}).Error; err != nil {
				return fmt.Errorf("archiving stale provider campaigns: %w", err)
			}
		}
		if err := archiveMissingProviderAccounts(tx, platform, syncedAt); err != nil {
			return err
		}

		for _, page := range snapshot.Pages {
			if err := storeRawPayload(tx, runID, "page", page.ExternalID, page.Raw, syncedAt); err != nil {
				return err
			}
		}
		for _, pixel := range snapshot.Pixels {
			if err := storeRawPayload(tx, runID, "pixel", pixel.ExternalID, pixel.Raw, syncedAt); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return providersync.ApplyResult{}, fmt.Errorf("applying provider snapshot: %w", err)
	}

	return result, nil
}

func archiveMissingProviderAccounts(tx *gorm.DB, platform ads.Platform, syncedAt time.Time) error {
	staleAccountIDs := []string{}
	if err := tx.Model(&accountRecord{}).
		Where(
			"platform_code = ? AND last_synced_at IS NOT NULL AND last_synced_at <> ?",
			string(platform),
			syncedAt,
		).
		Pluck("id", &staleAccountIDs).Error; err != nil {
		return fmt.Errorf("finding stale provider accounts: %w", err)
	}
	if len(staleAccountIDs) == 0 {
		return nil
	}
	if err := tx.Model(&campaignRecord{}).
		Where("account_id IN ? AND last_synced_at IS NOT NULL", staleAccountIDs).
		Updates(map[string]any{"status": string(ads.StatusArchived), "updated_at": syncedAt}).Error; err != nil {
		return fmt.Errorf("archiving campaigns for stale provider accounts: %w", err)
	}
	if err := tx.Model(&accountRecord{}).
		Where("id IN ?", staleAccountIDs).
		Updates(map[string]any{"status": string(ads.StatusArchived), "updated_at": syncedAt}).Error; err != nil {
		return fmt.Errorf("archiving stale provider accounts: %w", err)
	}

	return nil
}

// FinishRun records a terminal sync state and returns the updated run.
func (s *ProviderSyncStore) FinishRun(
	ctx context.Context,
	id string,
	status string,
	recordsProcessed int64,
	errorSummary string,
	finishedAt time.Time,
) (providersync.Run, error) {
	updates := map[string]any{
		"status": status, "records_processed": recordsProcessed,
		"error_summary": errorSummary, "finished_at": finishedAt, "updated_at": finishedAt,
	}
	result := s.db.WithContext(ctx).Model(&syncRunRecord{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return providersync.Run{}, fmt.Errorf("updating sync run: %w", result.Error)
	}
	if result.RowsAffected != 1 {
		return providersync.Run{}, providersync.ErrUnavailable
	}

	var record syncRunRecord
	if err := s.db.WithContext(ctx).Where("id = ?", id).Take(&record).Error; err != nil {
		return providersync.Run{}, fmt.Errorf("reading finished sync run: %w", err)
	}

	return syncRunToDomain(record), nil
}

// ListRuns returns recent provider runs newest first.
func (s *ProviderSyncStore) ListRuns(
	ctx context.Context,
	platform ads.Platform,
	limit int,
) ([]providersync.Run, error) {
	records := []syncRunRecord{}
	if err := s.db.WithContext(ctx).
		Where("platform_code = ?", string(platform)).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("querying sync runs: %w", err)
	}

	runs := make([]providersync.Run, 0, len(records))
	for _, record := range records {
		runs = append(runs, syncRunToDomain(record))
	}

	return runs, nil
}

func upsertProviderAccount(
	tx *gorm.DB,
	platform ads.Platform,
	account connector.ExternalAdAccount,
	syncedAt time.Time,
) (string, error) {
	var record accountRecord
	err := tx.Where("platform_code = ? AND external_id = ?", string(platform), account.ExternalID).Take(&record).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("reading provider account: %w", err)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		id, idErr := newRecordID()
		if idErr != nil {
			return "", idErr
		}
		record = accountRecord{
			ID: id, PlatformCode: string(platform), ExternalID: account.ExternalID,
			Name: account.Name, Currency: account.Currency, Timezone: account.Timezone,
			Status: account.Status, ProviderData: normalizedRaw(account.Raw), LastSyncedAt: &syncedAt,
		}
		if err := tx.Create(&record).Error; err != nil {
			return "", fmt.Errorf("creating provider account: %w", err)
		}

		return record.ID, nil
	}

	updates := map[string]any{
		"name": account.Name, "currency": account.Currency, "timezone": account.Timezone,
		"status": account.Status, "provider_data": normalizedRaw(account.Raw),
		"last_synced_at": syncedAt, "updated_at": syncedAt,
	}
	if err := tx.Model(&accountRecord{}).Where("id = ?", record.ID).Updates(updates).Error; err != nil {
		return "", fmt.Errorf("updating provider account: %w", err)
	}

	return record.ID, nil
}

func upsertProviderCampaign(
	tx *gorm.DB,
	accountID string,
	campaign connector.ExternalCampaign,
	syncedAt time.Time,
) error {
	var record campaignRecord
	err := tx.Where("account_id = ? AND external_id = ?", accountID, campaign.ExternalID).Take(&record).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("reading provider campaign: %w", err)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		id, idErr := newRecordID()
		if idErr != nil {
			return idErr
		}
		record = campaignRecord{
			ID: id, AccountID: accountID, ExternalID: campaign.ExternalID,
			Name: campaign.Name, Objective: campaign.Objective, Status: campaign.Status,
			ProviderData: normalizedRaw(campaign.Raw), LastSyncedAt: &syncedAt,
		}
		if err := tx.Create(&record).Error; err != nil {
			return fmt.Errorf("creating provider campaign: %w", err)
		}

		return nil
	}

	updates := map[string]any{
		"name": campaign.Name, "objective": campaign.Objective, "status": campaign.Status,
		"provider_data": normalizedRaw(campaign.Raw), "last_synced_at": syncedAt, "updated_at": syncedAt,
	}
	if err := tx.Model(&campaignRecord{}).Where("id = ?", record.ID).Updates(updates).Error; err != nil {
		return fmt.Errorf("updating provider campaign: %w", err)
	}

	return nil
}

func storeRawPayload(
	tx *gorm.DB,
	runID string,
	kind string,
	externalID string,
	payload json.RawMessage,
	receivedAt time.Time,
) error {
	id, err := newRecordID()
	if err != nil {
		return err
	}
	normalized := normalizedRaw(payload)
	hash := sha256.Sum256(append([]byte(kind+"\x00"+externalID+"\x00"), normalized...))
	record := rawPayloadRecord{
		ID: id, SyncRunID: runID, EntityKind: kind, ExternalID: externalID,
		Payload: normalized, PayloadHash: fmt.Sprintf("%x", hash[:]), ReceivedAt: receivedAt,
	}
	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "sync_run_id"}, {Name: "payload_hash"}},
		DoNothing: true,
	}).Create(&record).Error; err != nil {
		return fmt.Errorf("persisting raw provider payload: %w", err)
	}

	return nil
}

func normalizedRaw(payload json.RawMessage) json.RawMessage {
	if len(payload) == 0 {
		return json.RawMessage(`{}`)
	}

	return payload
}

func syncRunToDomain(record syncRunRecord) providersync.Run {
	return providersync.Run{
		ID: record.ID, Platform: ads.Platform(record.PlatformCode), Status: record.Status,
		RecordsProcessed: record.RecordsProcessed, ErrorSummary: record.ErrorSummary,
		StartedAt: record.StartedAt, FinishedAt: record.FinishedAt,
		CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
	}
}

func newRecordID() (string, error) {
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
