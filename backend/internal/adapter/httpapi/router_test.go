package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/application/health"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/config"
	"go.uber.org/zap"
)

type healthPingerStub struct {
	err error
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
			server := NewServer(testServerConfig(), service, zap.NewNop())
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
			server := NewServer(testServerConfig(), service, zap.NewNop())
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
