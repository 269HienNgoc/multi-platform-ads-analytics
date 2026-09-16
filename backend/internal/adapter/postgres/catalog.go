package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/application/catalog"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain/ads"
	"gorm.io/gorm"
)

var _ catalog.Store = (*CatalogStore)(nil)

type accountRecord struct {
	ID           string          `gorm:"column:id;type:uuid;primaryKey"`
	PlatformCode string          `gorm:"column:platform_code"`
	ExternalID   string          `gorm:"column:external_id"`
	Name         string          `gorm:"column:name"`
	Currency     string          `gorm:"column:currency"`
	Timezone     string          `gorm:"column:timezone"`
	Status       string          `gorm:"column:status"`
	ProviderData json.RawMessage `gorm:"column:provider_data;type:jsonb"`
	LastSyncedAt *time.Time      `gorm:"column:last_synced_at"`
	CreatedAt    time.Time       `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time       `gorm:"column:updated_at;autoUpdateTime"`
}

func (accountRecord) TableName() string { return "ad_accounts" }

type campaignRecord struct {
	ID           string          `gorm:"column:id;type:uuid;primaryKey"`
	AccountID    string          `gorm:"column:account_id;type:uuid"`
	ExternalID   string          `gorm:"column:external_id"`
	Name         string          `gorm:"column:name"`
	Objective    string          `gorm:"column:objective"`
	Status       string          `gorm:"column:status"`
	ProviderData json.RawMessage `gorm:"column:provider_data;type:jsonb"`
	LastSyncedAt *time.Time      `gorm:"column:last_synced_at"`
	CreatedAt    time.Time       `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time       `gorm:"column:updated_at;autoUpdateTime"`
}

func (campaignRecord) TableName() string { return "campaigns" }

type adGroupRecord struct {
	ID           string          `gorm:"column:id;type:uuid;primaryKey"`
	CampaignID   string          `gorm:"column:campaign_id;type:uuid"`
	ExternalID   string          `gorm:"column:external_id"`
	Name         string          `gorm:"column:name"`
	Status       string          `gorm:"column:status"`
	ProviderData json.RawMessage `gorm:"column:provider_data;type:jsonb"`
	CreatedAt    time.Time       `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time       `gorm:"column:updated_at;autoUpdateTime"`
}

func (adGroupRecord) TableName() string { return "ad_groups" }

type adRecord struct {
	ID           string          `gorm:"column:id;type:uuid;primaryKey"`
	AdGroupID    string          `gorm:"column:ad_group_id;type:uuid"`
	ExternalID   string          `gorm:"column:external_id"`
	Name         string          `gorm:"column:name"`
	Status       string          `gorm:"column:status"`
	ProviderData json.RawMessage `gorm:"column:provider_data;type:jsonb"`
	CreatedAt    time.Time       `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time       `gorm:"column:updated_at;autoUpdateTime"`
}

func (adRecord) TableName() string { return "ads" }

type creativeRecord struct {
	ID           string          `gorm:"column:id;type:uuid;primaryKey"`
	AdID         string          `gorm:"column:ad_id;type:uuid"`
	ExternalID   string          `gorm:"column:external_id"`
	Name         string          `gorm:"column:name"`
	Format       string          `gorm:"column:format"`
	AssetURL     string          `gorm:"column:asset_url"`
	ProviderData json.RawMessage `gorm:"column:provider_data;type:jsonb"`
	CreatedAt    time.Time       `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time       `gorm:"column:updated_at;autoUpdateTime"`
}

func (creativeRecord) TableName() string { return "creatives" }

// CatalogStore persists the normalized advertising hierarchy with GORM.
type CatalogStore struct {
	db *gorm.DB
}

// NewCatalogStore creates a PostgreSQL catalog adapter.
func NewCatalogStore(client *Client) *CatalogStore {
	return &CatalogStore{db: client.db}
}

// ListAccounts loads all accounts in a stable order.
func (s *CatalogStore) ListAccounts(ctx context.Context) ([]ads.AdAccount, error) {
	records := []accountRecord{}
	if err := s.db.WithContext(ctx).Order("platform_code, name, id").Find(&records).Error; err != nil {
		return nil, fmt.Errorf("querying ad accounts: %w", err)
	}
	type campaignCount struct {
		AccountID string `gorm:"column:account_id"`
		Count     int64  `gorm:"column:count"`
	}
	counts := []campaignCount{}
	if err := s.db.WithContext(ctx).
		Model(&campaignRecord{}).
		Select("account_id, COUNT(*) AS count").
		Group("account_id").
		Scan(&counts).Error; err != nil {
		return nil, fmt.Errorf("counting account campaigns: %w", err)
	}
	countsByAccount := make(map[string]int64, len(counts))
	for _, item := range counts {
		countsByAccount[item.AccountID] = item.Count
	}

	accounts := make([]ads.AdAccount, 0, len(records))
	for _, record := range records {
		account, err := accountToDomain(record)
		if err != nil {
			return nil, err
		}
		account.CampaignCount = countsByAccount[account.ID]
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// CreateAccount persists a new advertising account.
func (s *CatalogStore) CreateAccount(ctx context.Context, account *ads.AdAccount) error {
	providerData, err := marshalProviderData(account.ProviderData)
	if err != nil {
		return err
	}
	record := accountRecord{
		ID: account.ID, PlatformCode: string(account.Platform), ExternalID: account.ExternalID,
		Name: account.Name, Currency: account.Currency, Timezone: account.Timezone,
		Status: string(account.Status), ProviderData: providerData,
		LastSyncedAt: account.LastSyncedAt,
	}
	if err := s.db.WithContext(ctx).Create(&record).Error; err != nil {
		return translateWriteError(err)
	}
	account.CreatedAt = record.CreatedAt
	account.UpdatedAt = record.UpdatedAt

	return nil
}

// CreateCampaign persists a new campaign.
func (s *CatalogStore) CreateCampaign(ctx context.Context, campaign *ads.Campaign) error {
	providerData, err := marshalProviderData(campaign.ProviderData)
	if err != nil {
		return err
	}
	record := campaignRecord{
		ID: campaign.ID, AccountID: campaign.AccountID, ExternalID: campaign.ExternalID,
		Name: campaign.Name, Objective: campaign.Objective, Status: string(campaign.Status), ProviderData: providerData,
		LastSyncedAt: campaign.LastSyncedAt,
	}
	if err := s.db.WithContext(ctx).Create(&record).Error; err != nil {
		return translateWriteError(err)
	}
	campaign.CreatedAt = record.CreatedAt
	campaign.UpdatedAt = record.UpdatedAt

	return nil
}

// CreateAdGroup persists a new ad group.
func (s *CatalogStore) CreateAdGroup(ctx context.Context, group *ads.AdGroup) error {
	providerData, err := marshalProviderData(group.ProviderData)
	if err != nil {
		return err
	}
	record := adGroupRecord{
		ID: group.ID, CampaignID: group.CampaignID, ExternalID: group.ExternalID,
		Name: group.Name, Status: string(group.Status), ProviderData: providerData,
	}
	if err := s.db.WithContext(ctx).Create(&record).Error; err != nil {
		return translateWriteError(err)
	}
	group.CreatedAt = record.CreatedAt
	group.UpdatedAt = record.UpdatedAt

	return nil
}

// CreateAd persists a new advertisement.
func (s *CatalogStore) CreateAd(ctx context.Context, ad *ads.Ad) error {
	providerData, err := marshalProviderData(ad.ProviderData)
	if err != nil {
		return err
	}
	record := adRecord{
		ID: ad.ID, AdGroupID: ad.AdGroupID, ExternalID: ad.ExternalID,
		Name: ad.Name, Status: string(ad.Status), ProviderData: providerData,
	}
	if err := s.db.WithContext(ctx).Create(&record).Error; err != nil {
		return translateWriteError(err)
	}
	ad.CreatedAt = record.CreatedAt
	ad.UpdatedAt = record.UpdatedAt

	return nil
}

// CreateCreative persists a new creative.
func (s *CatalogStore) CreateCreative(ctx context.Context, creative *ads.Creative) error {
	providerData, err := marshalProviderData(creative.ProviderData)
	if err != nil {
		return err
	}
	record := creativeRecord{
		ID: creative.ID, AdID: creative.AdID, ExternalID: creative.ExternalID,
		Name: creative.Name, Format: creative.Format, AssetURL: creative.AssetURL, ProviderData: providerData,
	}
	if err := s.db.WithContext(ctx).Create(&record).Error; err != nil {
		return translateWriteError(err)
	}
	creative.CreatedAt = record.CreatedAt
	creative.UpdatedAt = record.UpdatedAt

	return nil
}

// AccountHierarchy loads one account tree with a bounded number of queries.
func (s *CatalogStore) AccountHierarchy(ctx context.Context, accountID string) (ads.AccountHierarchy, error) {
	var account accountRecord
	if err := s.db.WithContext(ctx).Where("id = ?", accountID).Take(&account).Error; err != nil {
		return ads.AccountHierarchy{}, translateReadError(err)
	}

	var campaigns []campaignRecord
	if err := s.db.WithContext(ctx).Where("account_id = ?", accountID).Order("created_at, id").Find(&campaigns).Error; err != nil {
		return ads.AccountHierarchy{}, fmt.Errorf("querying campaigns: %w", err)
	}

	campaignIDs := make([]string, 0, len(campaigns))
	for _, campaign := range campaigns {
		campaignIDs = append(campaignIDs, campaign.ID)
	}
	var groups []adGroupRecord
	if len(campaignIDs) > 0 {
		if err := s.db.WithContext(ctx).Where("campaign_id IN ?", campaignIDs).Order("created_at, id").Find(&groups).Error; err != nil {
			return ads.AccountHierarchy{}, fmt.Errorf("querying ad groups: %w", err)
		}
	}

	groupIDs := make([]string, 0, len(groups))
	for _, group := range groups {
		groupIDs = append(groupIDs, group.ID)
	}
	var adRecords []adRecord
	if len(groupIDs) > 0 {
		if err := s.db.WithContext(ctx).Where("ad_group_id IN ?", groupIDs).Order("created_at, id").Find(&adRecords).Error; err != nil {
			return ads.AccountHierarchy{}, fmt.Errorf("querying ads: %w", err)
		}
	}

	adIDs := make([]string, 0, len(adRecords))
	for _, ad := range adRecords {
		adIDs = append(adIDs, ad.ID)
	}
	var creatives []creativeRecord
	if len(adIDs) > 0 {
		if err := s.db.WithContext(ctx).Where("ad_id IN ?", adIDs).Order("created_at, id").Find(&creatives).Error; err != nil {
			return ads.AccountHierarchy{}, fmt.Errorf("querying creatives: %w", err)
		}
	}

	return assembleHierarchy(account, campaigns, groups, adRecords, creatives)
}

func assembleHierarchy(
	account accountRecord,
	campaigns []campaignRecord,
	groups []adGroupRecord,
	adRecords []adRecord,
	creatives []creativeRecord,
) (ads.AccountHierarchy, error) {
	domainAccount, err := accountToDomain(account)
	if err != nil {
		return ads.AccountHierarchy{}, err
	}
	hierarchy := ads.AccountHierarchy{Account: domainAccount, Campaigns: make([]ads.CampaignNode, 0, len(campaigns))}
	campaignIndex := make(map[string]int, len(campaigns))
	for _, record := range campaigns {
		campaign, conversionErr := campaignToDomain(record)
		if conversionErr != nil {
			return ads.AccountHierarchy{}, conversionErr
		}
		campaignIndex[campaign.ID] = len(hierarchy.Campaigns)
		hierarchy.Campaigns = append(hierarchy.Campaigns, ads.CampaignNode{Campaign: campaign, AdGroups: []ads.AdGroupNode{}})
	}

	type groupLocation struct{ campaign, group int }
	groupIndex := make(map[string]groupLocation, len(groups))
	for _, record := range groups {
		campaignPosition, exists := campaignIndex[record.CampaignID]
		if !exists {
			return ads.AccountHierarchy{}, errors.New("assembling hierarchy: ad group has unknown campaign")
		}
		group, conversionErr := adGroupToDomain(record)
		if conversionErr != nil {
			return ads.AccountHierarchy{}, conversionErr
		}
		groupPosition := len(hierarchy.Campaigns[campaignPosition].AdGroups)
		groupIndex[group.ID] = groupLocation{campaign: campaignPosition, group: groupPosition}
		hierarchy.Campaigns[campaignPosition].AdGroups = append(
			hierarchy.Campaigns[campaignPosition].AdGroups,
			ads.AdGroupNode{AdGroup: group, Ads: []ads.AdNode{}},
		)
	}

	type adLocation struct{ campaign, group, ad int }
	adIndex := make(map[string]adLocation, len(adRecords))
	for _, record := range adRecords {
		location, exists := groupIndex[record.AdGroupID]
		if !exists {
			return ads.AccountHierarchy{}, errors.New("assembling hierarchy: ad has unknown ad group")
		}
		ad, conversionErr := adToDomain(record)
		if conversionErr != nil {
			return ads.AccountHierarchy{}, conversionErr
		}
		adPosition := len(hierarchy.Campaigns[location.campaign].AdGroups[location.group].Ads)
		adIndex[ad.ID] = adLocation{campaign: location.campaign, group: location.group, ad: adPosition}
		hierarchy.Campaigns[location.campaign].AdGroups[location.group].Ads = append(
			hierarchy.Campaigns[location.campaign].AdGroups[location.group].Ads,
			ads.AdNode{Ad: ad, Creatives: []ads.Creative{}},
		)
	}

	for _, record := range creatives {
		location, exists := adIndex[record.AdID]
		if !exists {
			return ads.AccountHierarchy{}, errors.New("assembling hierarchy: creative has unknown ad")
		}
		creative, conversionErr := creativeToDomain(record)
		if conversionErr != nil {
			return ads.AccountHierarchy{}, conversionErr
		}
		node := &hierarchy.Campaigns[location.campaign].AdGroups[location.group].Ads[location.ad]
		node.Creatives = append(node.Creatives, creative)
	}

	return hierarchy, nil
}

func marshalProviderData(data ads.ProviderData) (json.RawMessage, error) {
	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("encoding provider data: %w", err)
	}

	return encoded, nil
}

func unmarshalProviderData(data json.RawMessage) (ads.ProviderData, error) {
	providerData := ads.ProviderData{}
	if len(data) == 0 {
		return providerData, nil
	}
	if err := json.Unmarshal(data, &providerData); err != nil {
		return nil, fmt.Errorf("decoding provider data: %w", err)
	}

	return providerData, nil
}

func accountToDomain(record accountRecord) (ads.AdAccount, error) {
	providerData, err := unmarshalProviderData(record.ProviderData)
	if err != nil {
		return ads.AdAccount{}, err
	}

	return ads.AdAccount{
		ID: record.ID, Platform: ads.Platform(record.PlatformCode), ExternalID: record.ExternalID,
		Name: record.Name, Currency: record.Currency, Timezone: record.Timezone,
		Status: ads.Status(record.Status), ProviderData: providerData,
		LastSyncedAt: record.LastSyncedAt,
		CreatedAt:    record.CreatedAt, UpdatedAt: record.UpdatedAt,
	}, nil
}

func campaignToDomain(record campaignRecord) (ads.Campaign, error) {
	providerData, err := unmarshalProviderData(record.ProviderData)
	if err != nil {
		return ads.Campaign{}, err
	}

	return ads.Campaign{
		ID: record.ID, AccountID: record.AccountID, ExternalID: record.ExternalID,
		Name: record.Name, Objective: record.Objective, Status: ads.Status(record.Status),
		ProviderData: providerData, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
		LastSyncedAt: record.LastSyncedAt,
	}, nil
}

func adGroupToDomain(record adGroupRecord) (ads.AdGroup, error) {
	providerData, err := unmarshalProviderData(record.ProviderData)
	if err != nil {
		return ads.AdGroup{}, err
	}

	return ads.AdGroup{
		ID: record.ID, CampaignID: record.CampaignID, ExternalID: record.ExternalID,
		Name: record.Name, Status: ads.Status(record.Status), ProviderData: providerData,
		CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
	}, nil
}

func adToDomain(record adRecord) (ads.Ad, error) {
	providerData, err := unmarshalProviderData(record.ProviderData)
	if err != nil {
		return ads.Ad{}, err
	}

	return ads.Ad{
		ID: record.ID, AdGroupID: record.AdGroupID, ExternalID: record.ExternalID,
		Name: record.Name, Status: ads.Status(record.Status), ProviderData: providerData,
		CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
	}, nil
}

func creativeToDomain(record creativeRecord) (ads.Creative, error) {
	providerData, err := unmarshalProviderData(record.ProviderData)
	if err != nil {
		return ads.Creative{}, err
	}

	return ads.Creative{
		ID: record.ID, AdID: record.AdID, ExternalID: record.ExternalID,
		Name: record.Name, Format: record.Format, AssetURL: record.AssetURL,
		ProviderData: providerData, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt,
	}, nil
}

func translateWriteError(err error) error {
	switch {
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return catalog.ErrConflict
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		return catalog.ErrNotFound
	default:
		return fmt.Errorf("persisting catalog entity: %w", err)
	}
}

func translateReadError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return catalog.ErrNotFound
	}

	return fmt.Errorf("reading catalog entity: %w", err)
}
