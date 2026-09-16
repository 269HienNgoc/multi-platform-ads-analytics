// Package connector defines provider boundaries consumed by advertising use cases.
package connector

import (
	"context"
	"encoding/json"

	automationdomain "github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain/automation"
)

// ExternalAsset is a provider-owned object normalized for the application layer.
type ExternalAsset struct {
	ExternalID string          `json:"external_id"`
	Name       string          `json:"name"`
	Status     string          `json:"status,omitempty"`
	Raw        json.RawMessage `json:"-"`
}

// ExternalAdAccount is one provider account normalized for catalog persistence.
type ExternalAdAccount struct {
	ExternalID string          `json:"external_id"`
	Name       string          `json:"name"`
	Currency   string          `json:"currency"`
	Timezone   string          `json:"timezone"`
	Status     string          `json:"status"`
	Raw        json.RawMessage `json:"-"`
}

// ExternalCampaign is one provider campaign and its owning external account.
type ExternalCampaign struct {
	AccountExternalID string          `json:"account_external_id"`
	ExternalID        string          `json:"external_id"`
	Name              string          `json:"name"`
	Objective         string          `json:"objective"`
	Status            string          `json:"status"`
	Raw               json.RawMessage `json:"-"`
}

// AssetSnapshot contains assets visible through one provider connection.
type AssetSnapshot struct {
	AdAccounts []ExternalAdAccount `json:"ad_accounts"`
	Campaigns  []ExternalCampaign  `json:"campaigns"`
	Pages      []ExternalAsset     `json:"pages"`
	Pixels     []ExternalAsset     `json:"pixels"`
	Warnings   []string            `json:"warnings,omitempty"`
}

// CreateSeedInput contains provider-neutral inputs for an engagement seed campaign.
type CreateSeedInput struct {
	AdAccountExternalID string
	PageExternalID      string
	ExistingPostID      string
	DailyBudgetMinor    int64
	Latitude            float64
	Longitude           float64
	RadiusKM            int
}

// CreateMainInput contains provider-neutral inputs for a conversion campaign.
type CreateMainInput struct {
	AdAccountExternalID string
	PageExternalID      string
	PixelExternalID     string
	PixelEvent          string
	DestinationURL      string
	CTA                 string
	DailyBudgetMinor    int64
}

// AdvertisingPlatform isolates workflow code from provider SDK and API models.
type AdvertisingPlatform interface {
	SyncAssets(context.Context) (AssetSnapshot, error)
	CreateSeedCampaign(context.Context, CreateSeedInput) (string, error)
	CreateMainCampaign(context.Context, CreateMainInput) (string, error)
	PauseCampaign(context.Context, string) error
	UpdateCampaignBudget(context.Context, string, int64) error
	GetCampaignMetrics(context.Context, string) (automationdomain.CampaignMetrics, error)
}
