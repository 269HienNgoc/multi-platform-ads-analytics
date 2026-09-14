package catalog

import (
	"context"
	"errors"
	"testing"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain/ads"
)

type storeStub struct {
	createAccountFn  func(context.Context, *ads.AdAccount) error
	createCampaignFn func(context.Context, *ads.Campaign) error
	createAdGroupFn  func(context.Context, *ads.AdGroup) error
	createAdFn       func(context.Context, *ads.Ad) error
	createCreativeFn func(context.Context, *ads.Creative) error
	hierarchyFn      func(context.Context, string) (ads.AccountHierarchy, error)
}

func (s storeStub) CreateAccount(ctx context.Context, account *ads.AdAccount) error {
	return s.createAccountFn(ctx, account)
}

func (s storeStub) CreateCampaign(ctx context.Context, campaign *ads.Campaign) error {
	if s.createCampaignFn == nil {
		return nil
	}

	return s.createCampaignFn(ctx, campaign)
}

func (s storeStub) CreateAdGroup(ctx context.Context, group *ads.AdGroup) error {
	if s.createAdGroupFn == nil {
		return nil
	}

	return s.createAdGroupFn(ctx, group)
}

func (s storeStub) CreateAd(ctx context.Context, ad *ads.Ad) error {
	if s.createAdFn == nil {
		return nil
	}

	return s.createAdFn(ctx, ad)
}

func (s storeStub) CreateCreative(ctx context.Context, creative *ads.Creative) error {
	if s.createCreativeFn == nil {
		return nil
	}

	return s.createCreativeFn(ctx, creative)
}

func (s storeStub) AccountHierarchy(ctx context.Context, accountID string) (ads.AccountHierarchy, error) {
	return s.hierarchyFn(ctx, accountID)
}

func TestService_CreateAccount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		account     ads.AdAccount
		storeErr    error
		expectedErr error
	}{
		{
			name: "valid account",
			account: ads.AdAccount{
				Platform: ads.PlatformMeta, ExternalID: "act_123", Name: " Main ", Currency: "usd",
				Timezone: "Asia/Ho_Chi_Minh", Status: ads.StatusActive,
			},
		},
		{name: "invalid platform", account: ads.AdAccount{}, expectedErr: ErrInvalidInput},
		{
			name: "store conflict",
			account: ads.AdAccount{
				Platform: ads.PlatformMeta, ExternalID: "act_123", Name: "Main", Currency: "USD",
				Timezone: "UTC", Status: ads.StatusActive,
			},
			storeErr: ErrConflict, expectedErr: ErrConflict,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			store := storeStub{createAccountFn: func(_ context.Context, account *ads.AdAccount) error {
				if !isUUID(account.ID) {
					t.Errorf("generated id is invalid: %q", account.ID)
				}
				if account.Currency != "USD" {
					t.Errorf("currency = %q, expected USD", account.Currency)
				}

				return test.storeErr
			}}
			service := New(store)
			account, err := service.CreateAccount(t.Context(), test.account)
			if !errors.Is(err, test.expectedErr) {
				t.Errorf("error = %v, expected %v", err, test.expectedErr)
			}
			if test.expectedErr == nil && account.Name != "Main" {
				t.Errorf("name = %q, expected Main", account.Name)
			}
		})
	}
}

func TestService_AccountHierarchy(t *testing.T) {
	t.Parallel()

	accountID := "00000000-0000-4000-8000-000000000001"
	expected := ads.AccountHierarchy{Account: ads.AdAccount{ID: accountID}}
	service := New(storeStub{
		createAccountFn: func(context.Context, *ads.AdAccount) error { return nil },
		hierarchyFn: func(_ context.Context, actualID string) (ads.AccountHierarchy, error) {
			if actualID != accountID {
				t.Errorf("account id = %q, expected %q", actualID, accountID)
			}

			return expected, nil
		},
	})

	actual, err := service.AccountHierarchy(t.Context(), accountID)
	if err != nil {
		t.Fatalf("AccountHierarchy() error = %v", err)
	}
	if actual.Account.ID != expected.Account.ID {
		t.Errorf("account id = %q, expected %q", actual.Account.ID, expected.Account.ID)
	}
}
