package ads

import "time"

// Status is the provider-neutral lifecycle state of an advertising entity.
type Status string

const (
	// StatusUnknown represents an unset lifecycle state.
	StatusUnknown Status = ""
	// StatusActive represents an entity that can currently deliver ads.
	StatusActive Status = "active"
	// StatusPaused represents an entity intentionally not delivering ads.
	StatusPaused Status = "paused"
	// StatusArchived represents an entity retained only for history.
	StatusArchived Status = "archived"
)

// IsValid reports whether the status belongs to the canonical set.
func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusPaused, StatusArchived:
		return true
	case StatusUnknown:
		return false
	default:
		return false
	}
}

// ProviderData stores provider-specific fields outside the normalized model.
type ProviderData map[string]any

// AdAccount is a provider-neutral advertising account.
type AdAccount struct {
	ID            string
	Platform      Platform
	ExternalID    string
	Name          string
	Currency      string
	Timezone      string
	Status        Status
	CampaignCount int64
	ProviderData  ProviderData
	LastSyncedAt  *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Campaign is a campaign owned by one advertising account.
type Campaign struct {
	ID           string
	AccountID    string
	ExternalID   string
	Name         string
	Objective    string
	Status       Status
	ProviderData ProviderData
	LastSyncedAt *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// AdGroup is a provider-neutral Meta ad set, Google ad group, or equivalent.
type AdGroup struct {
	ID           string
	CampaignID   string
	ExternalID   string
	Name         string
	Status       Status
	ProviderData ProviderData
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Ad is an advertisement owned by one ad group.
type Ad struct {
	ID           string
	AdGroupID    string
	ExternalID   string
	Name         string
	Status       Status
	ProviderData ProviderData
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Creative is the content attached to one advertisement.
type Creative struct {
	ID           string
	AdID         string
	ExternalID   string
	Name         string
	Format       string
	AssetURL     string
	ProviderData ProviderData
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// AccountHierarchy contains one account and all of its normalized descendants.
type AccountHierarchy struct {
	Account   AdAccount
	Campaigns []CampaignNode
}

// CampaignNode adds child ad groups to a campaign.
type CampaignNode struct {
	Campaign Campaign
	AdGroups []AdGroupNode
}

// AdGroupNode adds child ads to an ad group.
type AdGroupNode struct {
	AdGroup AdGroup
	Ads     []AdNode
}

// AdNode adds child creatives to an ad.
type AdNode struct {
	Ad        Ad
	Creatives []Creative
}
