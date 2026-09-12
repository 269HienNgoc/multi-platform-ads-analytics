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
}

func TestLoadEnvironmentOverride(t *testing.T) {
	configFile := writeConfig(t, validYAML())
	t.Setenv("ADS_DATABASE_HOST", "db.internal")
	t.Setenv("ADS_LOG_LEVEL", "warn")

	cfg, err := Load(configFile)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Database.Host != "db.internal" {
		t.Errorf("Database.Host = %q, expected %q", cfg.Database.Host, "db.internal")
	}
	if cfg.Log.Level != "warn" {
		t.Errorf("Log.Level = %q, expected %q", cfg.Log.Level, "warn")
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

func TestDatabaseDSN(t *testing.T) {
	t.Parallel()

	database := Database{
		Host:     "db.internal",
		Port:     5432,
		Name:     "ads data",
		User:     "ads-user",
		Password: "a secret with spaces",
		SSLMode:  "require",
		Timezone: "UTC",
	}

	dsn := database.DSN()
	if strings.Contains(dsn, "a secret with spaces") {
		t.Fatal("DSN() did not URL-encode the password")
	}
	if !strings.Contains(dsn, "sslmode=require") {
		t.Errorf("DSN() = %q, expected sslmode query", dsn)
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
database:
  host: localhost
  port: 5432
  name: ads
  user: postgres
  ssl_mode: disable
  timezone: UTC
  connect_timeout: 1s
  max_open_connections: 5
  max_idle_connections: 2
  connection_max_lifetime: 10m
  connection_max_idle_time: 2m
  slow_query_threshold: 100ms
log:
  level: debug
  encoding: console
`
}
