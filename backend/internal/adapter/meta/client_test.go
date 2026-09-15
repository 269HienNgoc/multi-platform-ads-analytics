package meta

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_SyncAssets(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer secret" {
			t.Error("authorization header was not set")
		}
		if request.URL.Path != "/v-test/me/adaccounts" {
			t.Errorf("path = %q, expected /v-test/me/adaccounts", request.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"data":[{"id":"act_1","name":"Main","account_status":1}]}`)); err != nil {
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
	if len(snapshot.AdAccounts) != 1 || snapshot.AdAccounts[0].ExternalID != "act_1" {
		t.Errorf("SyncAssets() = %#v, expected one account", snapshot)
	}
}
