package postgres

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain/ads"
)

func TestAssembleHierarchy(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	hierarchy, err := assembleHierarchy(
		accountRecord{ID: "account", PlatformCode: "meta", ProviderData: json.RawMessage(`{}`), CreatedAt: now, UpdatedAt: now},
		[]campaignRecord{{ID: "campaign", AccountID: "account", ProviderData: json.RawMessage(`{}`)}},
		[]adGroupRecord{{ID: "group", CampaignID: "campaign", ProviderData: json.RawMessage(`{}`)}},
		[]adRecord{{ID: "ad", AdGroupID: "group", ProviderData: json.RawMessage(`{}`)}},
		[]creativeRecord{{ID: "creative", AdID: "ad", ProviderData: json.RawMessage(`{"headline":"Hello"}`)}},
	)
	if err != nil {
		t.Fatalf("assembleHierarchy() error = %v", err)
	}
	if hierarchy.Account.Platform != ads.PlatformMeta {
		t.Errorf("platform = %q, expected %q", hierarchy.Account.Platform, ads.PlatformMeta)
	}
	creative := hierarchy.Campaigns[0].AdGroups[0].Ads[0].Creatives[0]
	if creative.ProviderData["headline"] != "Hello" {
		t.Errorf("headline = %v, expected Hello", creative.ProviderData["headline"])
	}
}

func TestUnmarshalProviderData(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       json.RawMessage
		expectsError bool
	}{
		{name: "empty", input: nil},
		{name: "object", input: json.RawMessage(`{"budget":10}`)},
		{name: "invalid", input: json.RawMessage(`[]`), expectsError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := unmarshalProviderData(test.input)
			if (err != nil) != test.expectsError {
				t.Errorf("error = %v, expects error = %t", err, test.expectsError)
			}
		})
	}
}
