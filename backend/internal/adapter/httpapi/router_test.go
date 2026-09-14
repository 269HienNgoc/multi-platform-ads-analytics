package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/application/catalog"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/application/health"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/config"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain/ads"
	"go.uber.org/zap"
)

type healthPingerStub struct {
	err error
}

type catalogServiceStub struct {
	createAccountFn func(context.Context, ads.AdAccount) (ads.AdAccount, error)
	hierarchyFn     func(context.Context, string) (ads.AccountHierarchy, error)
}

func (s catalogServiceStub) CreateAccount(ctx context.Context, account ads.AdAccount) (ads.AdAccount, error) {
	if s.createAccountFn == nil {
		return account, nil
	}

	return s.createAccountFn(ctx, account)
}

func (catalogServiceStub) CreateCampaign(_ context.Context, campaign ads.Campaign) (ads.Campaign, error) {
	return campaign, nil
}

func (catalogServiceStub) CreateAdGroup(_ context.Context, group ads.AdGroup) (ads.AdGroup, error) {
	return group, nil
}

func (catalogServiceStub) CreateAd(_ context.Context, ad ads.Ad) (ads.Ad, error) {
	return ad, nil
}

func (catalogServiceStub) CreateCreative(_ context.Context, creative ads.Creative) (ads.Creative, error) {
	return creative, nil
}

func (s catalogServiceStub) AccountHierarchy(ctx context.Context, accountID string) (ads.AccountHierarchy, error) {
	if s.hierarchyFn == nil {
		return ads.AccountHierarchy{}, nil
	}

	return s.hierarchyFn(ctx, accountID)
}

func (p healthPingerStub) Ping(context.Context) error {
	return p.err
}

func TestHealthRoutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		path           string
		pingErr        error
		expectedStatus int
	}{
		{name: "live", path: "/health/live", expectedStatus: http.StatusOK},
		{name: "ready", path: "/health/ready", expectedStatus: http.StatusOK},
		{
			name:           "database unavailable",
			path:           "/health/ready",
			pingErr:        errors.New("unavailable"),
			expectedStatus: http.StatusServiceUnavailable,
		},
		{name: "not found", path: "/missing", expectedStatus: http.StatusNotFound},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := health.New(healthPingerStub{err: test.pingErr})
			server := NewServer(testServerConfig(), service, catalogServiceStub{}, zap.NewNop())
			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()

			server.Handler.ServeHTTP(response, request)
			if response.Code != test.expectedStatus {
				t.Errorf("status = %d, expected %d", response.Code, test.expectedStatus)
			}
			if response.Header().Get(requestIDHeader) == "" {
				t.Error("response request id is empty")
			}
			if response.Header().Get("X-Content-Type-Options") != "nosniff" {
				t.Error("security headers were not applied")
			}
		})
	}
}

func TestRequestIDValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		requestID         string
		expectedRequestID string
		isGenerated       bool
	}{
		{name: "valid client id", requestID: "client_123", expectedRequestID: "client_123"},
		{name: "invalid client id", requestID: "bad id with spaces", isGenerated: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := health.New(healthPingerStub{})
			server := NewServer(testServerConfig(), service, catalogServiceStub{}, zap.NewNop())
			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/health/live", nil)
			request.Header.Set(requestIDHeader, test.requestID)
			response := httptest.NewRecorder()

			server.Handler.ServeHTTP(response, request)
			actualRequestID := response.Header().Get(requestIDHeader)
			if test.isGenerated {
				if actualRequestID == "" || actualRequestID == test.requestID {
					t.Errorf("request id = %q, expected a generated value", actualRequestID)
				}

				return
			}
			if actualRequestID != test.expectedRequestID {
				t.Errorf("request id = %q, expected %q", actualRequestID, test.expectedRequestID)
			}
		})
	}
}

func TestCatalogRoutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		service        catalogServiceStub
		expectedStatus int
	}{
		{
			name: "create account",
			method: http.MethodPost,
			path: "/api/v1/ad-accounts",
			body: `{"platform":"meta","external_id":"act_123","name":"Main","currency":"USD","timezone":"UTC","status":"active"}`,
			service: catalogServiceStub{createAccountFn: func(_ context.Context, account ads.AdAccount) (ads.AdAccount, error) {
				account.ID = "account-id"

				return account, nil
			}},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "invalid json",
			method: http.MethodPost,
			path: "/api/v1/ad-accounts",
			body: `{`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "duplicate account",
			method: http.MethodPost,
			path: "/api/v1/ad-accounts",
			body: `{"platform":"meta","external_id":"act_123","name":"Main","currency":"USD","timezone":"UTC","status":"active"}`,
			service: catalogServiceStub{createAccountFn: func(context.Context, ads.AdAccount) (ads.AdAccount, error) {
				return ads.AdAccount{}, catalog.ErrConflict
			}},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "account hierarchy not found",
			method: http.MethodGet,
			path: "/api/v1/ad-accounts/00000000-0000-0000-0000-000000000001/hierarchy",
			service: catalogServiceStub{hierarchyFn: func(context.Context, string) (ads.AccountHierarchy, error) {
				return ads.AccountHierarchy{}, catalog.ErrNotFound
			}},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			healthService := health.New(healthPingerStub{})
			server := NewServer(testServerConfig(), healthService, test.service, zap.NewNop())
			request := httptest.NewRequestWithContext(t.Context(), test.method, test.path, strings.NewReader(test.body))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()

			server.Handler.ServeHTTP(response, request)
			if response.Code != test.expectedStatus {
				t.Errorf("status = %d, expected %d; body = %s", response.Code, test.expectedStatus, response.Body.String())
			}
		})
	}
}

func testServerConfig() config.Server {
	return config.Server{
		Address:           "127.0.0.1:0",
		ReadHeaderTimeout: time.Second,
		ReadTimeout:       time.Second,
		WriteTimeout:      time.Second,
		IdleTimeout:       time.Second,
		ShutdownTimeout:   time.Second,
		MaxHeaderBytes:    1 << 20,
	}
}
