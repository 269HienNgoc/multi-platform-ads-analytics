package platform

import (
	"context"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain"
)

type AssetSnapshot struct {
	AdAccounts []ExternalAsset `json:"adAccounts"`
	Pages      []ExternalAsset `json:"pages"`
	Pixels     []ExternalAsset `json:"pixels"`
}

type ExternalAsset struct {
	ExternalID string `json:"externalId"`
	Name       string `json:"name"`
	Status     string `json:"status,omitempty"`
}

type CreateSeedInput struct {
	AdAccountExternalID string
	PageExternalID      string
	ExistingPostID      string
	DailyBudgetMinor    int64
	Latitude            float64
	Longitude           float64
	RadiusKM            int
}

type CreateMainInput struct {
	AdAccountExternalID string
	PageExternalID      string
	PixelExternalID     string
	PixelEvent          string
	DestinationURL      string
	CTA                  string
	DailyBudgetMinor    int64
}

// AdvertisingPlatform is the provider boundary. Core workflow code must depend
// on this interface rather than Meta/TikTok/Google-specific SDK models.
type AdvertisingPlatform interface {
	SyncAssets(ctx context.Context) (AssetSnapshot, error)
	CreateSeedCampaign(ctx context.Context, input CreateSeedInput) (string, error)
	CreateMainCampaign(ctx context.Context, input CreateMainInput) (string, error)
	PauseCampaign(ctx context.Context, externalCampaignID string) error
	UpdateCampaignBudget(ctx context.Context, externalCampaignID string, dailyBudgetMinor int64) error
	GetCampaignMetrics(ctx context.Context, externalCampaignID string) (domain.CampaignMetrics, error)
}
