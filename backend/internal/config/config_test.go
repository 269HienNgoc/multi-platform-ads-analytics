package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	configFile := writeConfig(t, validYAML())
	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.App.Name != "test-service" {
		t.Errorf("App.Name = %q, expected %q", cfg.App.Name, "test-service")
	}
	if cfg.Server.ReadTimeout != 2*time.Second {
		t.Errorf("Server.ReadTimeout = %v, expected %v", cfg.Server.ReadTimeout, 2*time.Second)
	}
	if cfg.Server.APIKey != "yaml-api-key" {
		t.Errorf("Server.APIKey = %q, expected YAML value", cfg.Server.APIKey)
	}
	if cfg.Database.Password != "yaml-password" {
		t.Errorf("Database.Password = %q, expected YAML value", cfg.Database.Password)
	}
	if cfg.Meta.AccessToken != "yaml-meta-token" {
		t.Errorf("Meta.AccessToken = %q, expected YAML value", cfg.Meta.AccessToken)
	}
}

func TestLoadIgnoresEnvironmentVariables(t *testing.T) {
	configFile := writeConfig(t, validYAML())
	t.Setenv("ADS_DATABASE_HOST", "db.internal")
	t.Setenv("ADS_LOG_LEVEL", "warn")

	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Database.Host != "localhost" {
		t.Errorf("Database.Host = %q, expected YAML value %q", cfg.Database.Host, "localhost")
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("Log.Level = %q, expected YAML value %q", cfg.Log.Level, "debug")
	}
}

func TestLoadRejectsInvalidConfig(t *testing.T) {
	t.Parallel()

	configFile := writeConfig(t, strings.Replace(validYAML(), "port: 5432", "port: 70000", 1))
	_, err := Load(configFile)
	if err == nil {
		t.Fatal("Load() error = nil, expected validation error")
	}
}

func TestLoadDoesNotSupplyHiddenDefaults(t *testing.T) {
	t.Parallel()

	content := strings.Replace(validYAML(), "  read_timeout: 2s\n", "", 1)
	_, err := Load(writeConfig(t, content))
	if err == nil || !strings.Contains(err.Error(), "server timeouts") {
		t.Fatalf("Load() error = %v, expected missing YAML value validation", err)
	}
}

func TestLoadRequiresAPIKeyInProduction(t *testing.T) {
	t.Parallel()

	content := strings.Replace(validYAML(), "environment: test", "environment: production", 1)
	content = strings.Replace(content, "api_key: yaml-api-key", `api_key: ""`, 1)
	_, err := Load(writeConfig(t, content))
	if err == nil || !strings.Contains(err.Error(), "api key") {
		t.Fatalf("Load() error = %v, expected production API key validation", err)
	}
}

func TestLoadRejectsEnabledMetaWithoutToken(t *testing.T) {
	t.Parallel()

	content := strings.Replace(validYAML(), "access_token: yaml-meta-token", `access_token: ""`, 1)
	_, err := Load(writeConfig(t, content))
	if err == nil || !strings.Contains(err.Error(), "access token") {
		t.Fatalf("Load() error = %v, expected Meta token validation", err)
	}
}

func TestDatabaseDSN(t *testing.T) {
	t.Parallel()

	database := Database{
		Host:           "db.internal",
		Port:           5432,
		Name:           "ads data",
		User:           "ads-user",
		Password:       "a secret with spaces",
		SSLMode:        "require",
		Timezone:       "UTC",
		ConnectTimeout: time.Second,
	}

	dsn := database.DSN()
	if strings.Contains(dsn, "a secret with spaces") {
		t.Fatal("DSN() did not URL-encode the password")
	}
	if !strings.Contains(dsn, "sslmode=require") {
		t.Errorf("DSN() = %q, expected sslmode query", dsn)
	}
	if !strings.Contains(dsn, "connect_timeout=1") {
		t.Errorf("DSN() = %q, expected connect timeout query", dsn)
	}
}

func writeConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing config: %v", err)
	}

	return path
}

func validYAML() string {
	return `app:
  name: test-service
  environment: test
server:
  address: 127.0.0.1:8080
  read_header_timeout: 1s
  read_timeout: 2s
  write_timeout: 3s
  idle_timeout: 4s
  shutdown_timeout: 5s
  max_header_bytes: 1024
  api_key: yaml-api-key
database:
  host: localhost
  port: 5432
  name: ads
  user: postgres
  password: yaml-password
  ssl_mode: disable
  timezone: UTC
  connect_timeout: 1s
  max_open_connections: 5
  max_idle_connections: 2
  connection_max_lifetime: 10m
  connection_max_idle_time: 2m
  slow_query_threshold: 100ms
meta:
  enabled: true
  base_url: https://graph.facebook.com
  version: v-test
  access_token: yaml-meta-token
  timeout: 1s
  sync_interval: 1m
log:
  level: debug
  encoding: console
`
}
