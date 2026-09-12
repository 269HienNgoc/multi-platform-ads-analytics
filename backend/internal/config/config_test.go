package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadYAMLConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte(`
server:
  address: ":9090"
meta:
  graph_version: "v23.0"
logging:
  level: debug
  encoding: console
  development: true
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.Server.Address != ":9090" {
		t.Fatalf("expected :9090, got %s", cfg.Server.Address)
	}
	if cfg.Logging.Level != "debug" {
		t.Fatalf("expected debug logging, got %s", cfg.Logging.Level)
	}
	if cfg.Database.DSN != "" {
		t.Fatalf("database dsn should remain optional in foundation")
	}
}

func TestLoadRejectsInvalidLoggingEncoding(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte(`
meta:
  graph_version: "v23.0"
logging:
  encoding: xml
`)

	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("expected invalid logging encoding error")
	}
}
