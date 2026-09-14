// Package catalog coordinates provider-neutral advertising catalog use cases.
package catalog

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain/ads"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// ErrInvalidInput identifies a request that violates catalog invariants.
var ErrInvalidInput = errors.New("catalog: invalid input")

// ErrNotFound identifies a requested catalog entity that does not exist.
var ErrNotFound = errors.New("catalog: not found")

// ErrConflict identifies a duplicate external identifier or another uniqueness conflict.
var ErrConflict = errors.New("catalog: conflict")

// Store is the persistence contract consumed by the catalog use cases.
type Store interface {
	CreateAccount(context.Context, *ads.AdAccount) error
	CreateCampaign(context.Context, *ads.Campaign) error
	CreateAdGroup(context.Context, *ads.AdGroup) error
	CreateAd(context.Context, *ads.Ad) error
	CreateCreative(context.Context, *ads.Creative) error
	AccountHierarchy(context.Context, string) (ads.AccountHierarchy, error)
}

// Service implements catalog use cases independently of transport and persistence.
type Service struct {
	store Store
}

// New creates a catalog service.
func New(store Store) *Service {
	return &Service{store: store}
}

// CreateAccount validates and persists an advertising account.
func (s *Service) CreateAccount(ctx context.Context, account ads.AdAccount) (ads.AdAccount, error) {
	id, err := newUUID()
	if err != nil {
		return ads.AdAccount{}, fmt.Errorf("generating account id: %w", err)
	}
	account.ID = id
	account.ExternalID = strings.TrimSpace(account.ExternalID)
	account.Name = strings.TrimSpace(account.Name)
	account.Currency = strings.ToUpper(strings.TrimSpace(account.Currency))
	account.Timezone = strings.TrimSpace(account.Timezone)
	account.ProviderData = normalizedProviderData(account.ProviderData)

	if !account.Platform.IsValid() {
		return ads.AdAccount{}, invalidField("platform")
	}
	if account.ExternalID == "" {
		return ads.AdAccount{}, invalidField("external_id")
	}
	if account.Name == "" {
		return ads.AdAccount{}, invalidField("name")
	}
	if len(account.Currency) != 3 {
		return ads.AdAccount{}, invalidField("currency")
	}
	if account.Timezone == "" {
		return ads.AdAccount{}, invalidField("timezone")
	}
	if !account.Status.IsValid() {
		return ads.AdAccount{}, invalidField("status")
	}
	if err := s.store.CreateAccount(ctx, &account); err != nil {
		return ads.AdAccount{}, fmt.Errorf("creating account: %w", err)
	}

	return account, nil
}

// CreateCampaign validates and persists a campaign.
func (s *Service) CreateCampaign(ctx context.Context, campaign ads.Campaign) (ads.Campaign, error) {
	id, err := newUUID()
	if err != nil {
		return ads.Campaign{}, fmt.Errorf("generating campaign id: %w", err)
	}
	campaign.ID = id
	campaign.AccountID = strings.TrimSpace(campaign.AccountID)
	campaign.ExternalID = strings.TrimSpace(campaign.ExternalID)
	campaign.Name = strings.TrimSpace(campaign.Name)
	campaign.Objective = strings.TrimSpace(campaign.Objective)
	campaign.ProviderData = normalizedProviderData(campaign.ProviderData)

	if !isUUID(campaign.AccountID) {
		return ads.Campaign{}, invalidField("account_id")
	}
	if campaign.ExternalID == "" {
		return ads.Campaign{}, invalidField("external_id")
	}
	if campaign.Name == "" {
		return ads.Campaign{}, invalidField("name")
	}
	if !campaign.Status.IsValid() {
		return ads.Campaign{}, invalidField("status")
	}
	if err := s.store.CreateCampaign(ctx, &campaign); err != nil {
		return ads.Campaign{}, fmt.Errorf("creating campaign: %w", err)
	}

	return campaign, nil
}

// CreateAdGroup validates and persists an ad group.
func (s *Service) CreateAdGroup(ctx context.Context, group ads.AdGroup) (ads.AdGroup, error) {
	id, err := newUUID()
	if err != nil {
		return ads.AdGroup{}, fmt.Errorf("generating ad group id: %w", err)
	}
	group.ID = id
	group.CampaignID = strings.TrimSpace(group.CampaignID)
	group.ExternalID = strings.TrimSpace(group.ExternalID)
	group.Name = strings.TrimSpace(group.Name)
	group.ProviderData = normalizedProviderData(group.ProviderData)

	if !isUUID(group.CampaignID) {
		return ads.AdGroup{}, invalidField("campaign_id")
	}
	if group.ExternalID == "" {
		return ads.AdGroup{}, invalidField("external_id")
	}
	if group.Name == "" {
		return ads.AdGroup{}, invalidField("name")
	}
	if !group.Status.IsValid() {
		return ads.AdGroup{}, invalidField("status")
	}
	if err := s.store.CreateAdGroup(ctx, &group); err != nil {
		return ads.AdGroup{}, fmt.Errorf("creating ad group: %w", err)
	}

	return group, nil
}

// CreateAd validates and persists an advertisement.
func (s *Service) CreateAd(ctx context.Context, ad ads.Ad) (ads.Ad, error) {
	id, err := newUUID()
	if err != nil {
		return ads.Ad{}, fmt.Errorf("generating ad id: %w", err)
	}
	ad.ID = id
	ad.AdGroupID = strings.TrimSpace(ad.AdGroupID)
	ad.ExternalID = strings.TrimSpace(ad.ExternalID)
	ad.Name = strings.TrimSpace(ad.Name)
	ad.ProviderData = normalizedProviderData(ad.ProviderData)

	if !isUUID(ad.AdGroupID) {
		return ads.Ad{}, invalidField("ad_group_id")
	}
	if ad.ExternalID == "" {
		return ads.Ad{}, invalidField("external_id")
	}
	if ad.Name == "" {
		return ads.Ad{}, invalidField("name")
	}
	if !ad.Status.IsValid() {
		return ads.Ad{}, invalidField("status")
	}
	if err := s.store.CreateAd(ctx, &ad); err != nil {
		return ads.Ad{}, fmt.Errorf("creating ad: %w", err)
	}

	return ad, nil
}

// CreateCreative validates and persists a creative.
func (s *Service) CreateCreative(ctx context.Context, creative ads.Creative) (ads.Creative, error) {
	id, err := newUUID()
	if err != nil {
		return ads.Creative{}, fmt.Errorf("generating creative id: %w", err)
	}
	creative.ID = id
	creative.AdID = strings.TrimSpace(creative.AdID)
	creative.ExternalID = strings.TrimSpace(creative.ExternalID)
	creative.Name = strings.TrimSpace(creative.Name)
	creative.Format = strings.ToLower(strings.TrimSpace(creative.Format))
	creative.AssetURL = strings.TrimSpace(creative.AssetURL)
	creative.ProviderData = normalizedProviderData(creative.ProviderData)

	if !isUUID(creative.AdID) {
		return ads.Creative{}, invalidField("ad_id")
	}
	if creative.ExternalID == "" {
		return ads.Creative{}, invalidField("external_id")
	}
	if creative.Name == "" {
		return ads.Creative{}, invalidField("name")
	}
	if creative.Format == "" {
		return ads.Creative{}, invalidField("format")
	}
	if err := s.store.CreateCreative(ctx, &creative); err != nil {
		return ads.Creative{}, fmt.Errorf("creating creative: %w", err)
	}

	return creative, nil
}

// AccountHierarchy returns one complete account subtree.
func (s *Service) AccountHierarchy(ctx context.Context, accountID string) (ads.AccountHierarchy, error) {
	accountID = strings.TrimSpace(accountID)
	if !isUUID(accountID) {
		return ads.AccountHierarchy{}, invalidField("account_id")
	}

	hierarchy, err := s.store.AccountHierarchy(ctx, accountID)
	if err != nil {
		return ads.AccountHierarchy{}, fmt.Errorf("loading account hierarchy: %w", err)
	}

	return hierarchy, nil
}

func normalizedProviderData(data ads.ProviderData) ads.ProviderData {
	if data == nil {
		return ads.ProviderData{}
	}

	return data
}

func invalidField(field string) error {
	return fmt.Errorf("%w: %s", ErrInvalidInput, field)
}

func isUUID(value string) bool {
	return uuidPattern.MatchString(strings.ToLower(value))
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
