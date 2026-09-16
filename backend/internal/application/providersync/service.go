// Package providersync coordinates durable imports from advertising providers.
package providersync

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/application/connector"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain/ads"
)

const (
	// StatusPending indicates a synchronization attempt waiting to run.
	StatusPending = "pending"
	// StatusRunning indicates a synchronization attempt currently importing data.
	StatusRunning = "running"
	// StatusSucceeded indicates a synchronization attempt completed successfully.
	StatusSucceeded = "succeeded"
	// StatusFailed indicates a synchronization attempt ended with an error.
	StatusFailed = "failed"
)

// ErrUnavailable indicates that a provider connector is not configured.
var ErrUnavailable = errors.New("provider sync: connector unavailable")

// Run is one observable provider synchronization attempt.
type Run struct {
	ID               string       `json:"id"`
	Platform         ads.Platform `json:"platform"`
	Status           string       `json:"status"`
	RecordsProcessed int64        `json:"records_processed"`
	ErrorSummary     string       `json:"error_summary,omitempty"`
	StartedAt        *time.Time   `json:"started_at,omitempty"`
	FinishedAt       *time.Time   `json:"finished_at,omitempty"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
}

// ApplyResult reports the catalog rows observed during a sync.
type ApplyResult struct {
	Accounts  int      `json:"accounts"`
	Campaigns int      `json:"campaigns"`
	Pages     int      `json:"pages"`
	Pixels    int      `json:"pixels"`
	Warnings  []string `json:"warnings,omitempty"`
}

// Store persists sync history and applies provider snapshots idempotently.
type Store interface {
	CreateRun(context.Context, Run) error
	ApplySnapshot(context.Context, string, ads.Platform, connector.AssetSnapshot, time.Time) (ApplyResult, error)
	FinishRun(context.Context, string, string, int64, string, time.Time) (Run, error)
	ListRuns(context.Context, ads.Platform, int) ([]Run, error)
}

// Provider is the read-only connector contract needed by synchronization.
type Provider interface {
	SyncAssets(context.Context) (connector.AssetSnapshot, error)
}

// Service synchronizes one advertising provider into the canonical catalog.
type Service struct {
	platform ads.Platform
	provider Provider
	store    Store
	now      func() time.Time
	gate     chan struct{}
}

// New creates a provider synchronization service.
func New(platform ads.Platform, provider Provider, store Store) *Service {
	return &Service{
		platform: platform,
		provider: provider,
		store:    store,
		now:      func() time.Time { return time.Now().UTC() },
		gate:     make(chan struct{}, 1),
	}
}

// Sync imports all assets visible to the configured provider connection.
func (s *Service) Sync(ctx context.Context) (Run, ApplyResult, error) {
	select {
	case s.gate <- struct{}{}:
		defer func() { <-s.gate }()
	case <-ctx.Done():
		return Run{}, ApplyResult{}, ctx.Err()
	}

	if s.provider == nil {
		return Run{}, ApplyResult{}, ErrUnavailable
	}

	startedAt := s.now()
	runID, err := newUUID()
	if err != nil {
		return Run{}, ApplyResult{}, fmt.Errorf("generating sync run id: %w", err)
	}
	run := Run{
		ID: runID, Platform: s.platform, Status: StatusRunning,
		StartedAt: &startedAt, CreatedAt: startedAt, UpdatedAt: startedAt,
	}
	if err := s.store.CreateRun(ctx, run); err != nil {
		return Run{}, ApplyResult{}, fmt.Errorf("creating sync run: %w", err)
	}

	snapshot, syncErr := s.provider.SyncAssets(ctx)
	if syncErr != nil {
		return s.fail(ctx, run, syncErr)
	}
	if err := validateSnapshot(snapshot); err != nil {
		return s.fail(ctx, run, err)
	}
	result, applyErr := s.store.ApplySnapshot(ctx, run.ID, s.platform, snapshot, startedAt)
	if applyErr != nil {
		return s.fail(ctx, run, applyErr)
	}
	result.Warnings = append([]string(nil), snapshot.Warnings...)

	processed := int64(result.Accounts + result.Campaigns + result.Pages + result.Pixels)
	finishedAt := s.now()
	completed, err := s.finishRun(
		ctx,
		run.ID,
		StatusSucceeded,
		processed,
		sanitizeWarnings(snapshot.Warnings),
		finishedAt,
	)
	if err != nil {
		return Run{}, ApplyResult{}, fmt.Errorf("finishing sync run: %w", err)
	}

	return completed, result, nil
}

// ListRuns returns recent sync history for the service provider.
func (s *Service) ListRuns(ctx context.Context, limit int) ([]Run, error) {
	if limit < 1 || limit > 100 {
		limit = 20
	}
	runs, err := s.store.ListRuns(ctx, s.platform, limit)
	if err != nil {
		return nil, fmt.Errorf("listing sync runs: %w", err)
	}
	if runs == nil {
		return []Run{}, nil
	}

	return runs, nil
}

func (s *Service) fail(ctx context.Context, run Run, cause error) (Run, ApplyResult, error) {
	finishedAt := s.now()
	summary := sanitizeError(cause)
	failed, finishErr := s.finishRun(ctx, run.ID, StatusFailed, 0, summary, finishedAt)
	if finishErr != nil {
		return Run{}, ApplyResult{}, errors.Join(
			fmt.Errorf("syncing %s assets: %w", s.platform, cause),
			fmt.Errorf("recording failed sync run: %w", finishErr),
		)
	}

	return failed, ApplyResult{}, fmt.Errorf("syncing %s assets: %w", s.platform, cause)
}

func (s *Service) finishRun(
	ctx context.Context,
	id string,
	status string,
	recordsProcessed int64,
	summary string,
	finishedAt time.Time,
) (Run, error) {
	finishCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()

	return s.store.FinishRun(finishCtx, id, status, recordsProcessed, summary, finishedAt)
}

func validateSnapshot(snapshot connector.AssetSnapshot) error {
	accounts := make(map[string]struct{}, len(snapshot.AdAccounts))
	for _, account := range snapshot.AdAccounts {
		externalID := strings.TrimSpace(account.ExternalID)
		if externalID == "" || len(externalID) > 255 || strings.TrimSpace(account.Name) == "" ||
			len(strings.TrimSpace(account.Name)) > 255 ||
			len(strings.TrimSpace(account.Currency)) != 3 || strings.TrimSpace(account.Timezone) == "" ||
			len(strings.TrimSpace(account.Timezone)) > 100 ||
			!ads.Status(account.Status).IsValid() {
			return fmt.Errorf("provider returned an invalid account %q", externalID)
		}
		if _, exists := accounts[externalID]; exists {
			return fmt.Errorf("provider returned duplicate account %q", externalID)
		}
		accounts[externalID] = struct{}{}
	}

	campaigns := make(map[string]struct{}, len(snapshot.Campaigns))
	for _, campaign := range snapshot.Campaigns {
		accountID := strings.TrimSpace(campaign.AccountExternalID)
		externalID := strings.TrimSpace(campaign.ExternalID)
		if _, exists := accounts[accountID]; !exists {
			return fmt.Errorf("provider campaign %q references unknown account %q", externalID, accountID)
		}
		if externalID == "" || len(externalID) > 255 || strings.TrimSpace(campaign.Name) == "" ||
			len(strings.TrimSpace(campaign.Name)) > 255 || len(strings.TrimSpace(campaign.Objective)) > 100 ||
			!ads.Status(campaign.Status).IsValid() {
			return fmt.Errorf("provider returned an invalid campaign %q", externalID)
		}
		key := accountID + "\x00" + externalID
		if _, exists := campaigns[key]; exists {
			return fmt.Errorf("provider returned duplicate campaign %q for account %q", externalID, accountID)
		}
		campaigns[key] = struct{}{}
	}
	if err := validateAssets("page", snapshot.Pages); err != nil {
		return err
	}
	if err := validateAssets("pixel", snapshot.Pixels); err != nil {
		return err
	}

	return nil
}

func validateAssets(kind string, assets []connector.ExternalAsset) error {
	for _, asset := range assets {
		externalID := strings.TrimSpace(asset.ExternalID)
		if externalID == "" || len(externalID) > 255 {
			return fmt.Errorf("provider returned an invalid %s %q", kind, externalID)
		}
	}

	return nil
}

func sanitizeError(err error) string {
	const maxLength = 1000

	message := strings.TrimSpace(err.Error())
	if len(message) > maxLength {
		return message[:maxLength]
	}

	return message
}

func sanitizeWarnings(warnings []string) string {
	if len(warnings) == 0 {
		return ""
	}

	return sanitizeError(errors.New(strings.Join(warnings, "; ")))
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
