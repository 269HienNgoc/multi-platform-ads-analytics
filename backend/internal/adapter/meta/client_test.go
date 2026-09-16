package meta

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestClient_SyncAssets(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer secret" {
			t.Error("authorization header was not set")
		}
		w.Header().Set("Content-Type", "application/json")
		var response string
		switch request.URL.Path {
		case "/v-test/me/adaccounts":
			if request.URL.Query().Get("after") == "account-page-2" {
				response = `{"data":[{"id":"act_2","name":"Backup","account_status":2,"currency":"USD","timezone_name":"UTC"}]}`
			} else {
				response = `{"data":[{"id":"act_1","name":"Main","account_status":1,"currency":"USD","timezone_name":"UTC"}],"paging":{"cursors":{"after":"account-page-2"},"next":"next"}}`
			}
		case "/v-test/act_1/campaigns":
			response = `{"data":[{"id":"campaign_1","name":"Awareness","objective":"OUTCOME_AWARENESS","status":"ACTIVE"}]}`
		case "/v-test/act_2/campaigns":
			response = `{"data":[{"id":"campaign_2","name":"Sales","objective":"OUTCOME_SALES","status":"PAUSED"}]}`
		case "/v-test/act_1/adspixels", "/v-test/act_2/adspixels":
			response = `{"data":[]}`
		case "/v-test/me/accounts":
			response = `{"data":[{"id":"page_1","name":"Main page"}]}`
		default:
			http.NotFound(w, request)

			return
		}
		if _, err := w.Write([]byte(response)); err != nil {
			t.Errorf("writing response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client, err := New(Config{BaseURL: server.URL, Version: "v-test", AccessToken: "secret", Timeout: time.Second})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	snapshot, err := client.SyncAssets(t.Context())
	if err != nil {
		t.Fatalf("SyncAssets() error = %v", err)
	}
	if len(snapshot.AdAccounts) != 2 || snapshot.AdAccounts[1].ExternalID != "act_2" {
		t.Errorf("SyncAssets() = %#v, expected two paginated accounts", snapshot)
	}
	if len(snapshot.Campaigns) != 2 || snapshot.Campaigns[1].AccountExternalID != "act_2" {
		t.Errorf("SyncAssets() = %#v, expected campaigns for both accounts", snapshot)
	}
	if len(snapshot.Pages) != 1 || len(snapshot.Warnings) != 0 {
		t.Errorf("SyncAssets() = %#v, expected page without warnings", snapshot)
	}
}

func TestClientSyncAssetsRetriesRateLimit(t *testing.T) {
	t.Parallel()

	var accountAttempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if request.URL.Path == "/v-test/me/adaccounts" && accountAttempts.Add(1) == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"type":"rate_limit","code":4}}`))

			return
		}
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	t.Cleanup(server.Close)

	client, err := New(Config{
		BaseURL: server.URL, Version: "v-test", AccessToken: "secret", Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := client.SyncAssets(t.Context()); err != nil {
		t.Fatalf("SyncAssets() error = %v", err)
	}
	if accountAttempts.Load() != 2 {
		t.Errorf("account attempts = %d, expected 2", accountAttempts.Load())
	}
}
