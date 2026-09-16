package providersync

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/application/connector"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain/ads"
)

type providerStub struct {
	snapshot connector.AssetSnapshot
	err      error
}

func (p providerStub) SyncAssets(context.Context) (connector.AssetSnapshot, error) {
	return p.snapshot, p.err
}

type storeStub struct {
	run      Run
	result   ApplyResult
	finished bool
}

func (s *storeStub) CreateRun(_ context.Context, run Run) error {
	s.run = run

	return nil
}

func (s *storeStub) ApplySnapshot(
	_ context.Context,
	_ string,
	_ ads.Platform,
	_ connector.AssetSnapshot,
	_ time.Time,
) (ApplyResult, error) {
	return s.result, nil
}

func (s *storeStub) FinishRun(
	_ context.Context,
	_ string,
	status string,
	recordsProcessed int64,
	errorSummary string,
	finishedAt time.Time,
) (Run, error) {
	s.finished = true
	s.run.Status = status
	s.run.RecordsProcessed = recordsProcessed
	s.run.ErrorSummary = errorSummary
	s.run.FinishedAt = &finishedAt

	return s.run, nil
}

func (s *storeStub) ListRuns(context.Context, ads.Platform, int) ([]Run, error) {
	return []Run{s.run}, nil
}

func TestServiceSyncPersistsWarnings(t *testing.T) {
	t.Parallel()

	snapshot := connector.AssetSnapshot{
		AdAccounts: []connector.ExternalAdAccount{{
			ExternalID: "act_1", Name: "Main", Currency: "USD", Timezone: "UTC", Status: "active",
		}},
		Campaigns: []connector.ExternalCampaign{{
			AccountExternalID: "act_1", ExternalID: "campaign_1", Name: "Sales", Status: "paused",
		}},
		Warnings: []string{"pages were not synchronized"},
	}
	store := &storeStub{result: ApplyResult{Accounts: 1, Campaigns: 1}}
	service := New(ads.PlatformMeta, providerStub{snapshot: snapshot}, store)
	service.now = func() time.Time { return time.Unix(100, 0).UTC() }

	run, result, err := service.Sync(t.Context())
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if run.Status != StatusSucceeded || run.RecordsProcessed != 2 || !store.finished {
		t.Errorf("Sync() run = %#v, expected completed run", run)
	}
	if len(result.Warnings) != 1 || run.ErrorSummary == "" {
		t.Errorf("Sync() result = %#v, run = %#v, expected persisted warning", result, run)
	}
}

func TestServiceSyncRecordsProviderFailure(t *testing.T) {
	t.Parallel()

	store := &storeStub{}
	service := New(ads.PlatformMeta, providerStub{err: errors.New("token expired")}, store)

	run, _, err := service.Sync(t.Context())
	if err == nil {
		t.Fatal("Sync() error = nil, expected provider failure")
	}
	if run.Status != StatusFailed || run.ErrorSummary != "token expired" {
		t.Errorf("Sync() run = %#v, expected failed run", run)
	}
}
