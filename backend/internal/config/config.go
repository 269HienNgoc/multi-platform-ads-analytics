package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App      AppConfig      `yaml:"app"`
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Meta     MetaConfig     `yaml:"meta"`
	Logging  LoggingConfig  `yaml:"logging"`
}

type AppConfig struct {
	Name        string `yaml:"name"`
	Environment string `yaml:"environment"`
}

type ServerConfig struct {
	Address                  string `yaml:"address"`
	ReadHeaderTimeoutSeconds int    `yaml:"read_header_timeout_seconds"`
	ShutdownTimeoutSeconds   int    `yaml:"shutdown_timeout_seconds"`
}

type DatabaseConfig struct {
	DSN string `yaml:"dsn"`
}

type RedisConfig struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	Database int    `yaml:"database"`
}

type MetaConfig struct {
	GraphBaseURL string `yaml:"graph_base_url"`
	GraphVersion string `yaml:"graph_version"`
	AccessToken  string `yaml:"access_token"`
}

type LoggingConfig struct {
	Level            string   `yaml:"level"`
	Encoding         string   `yaml:"encoding"`
	Development      bool     `yaml:"development"`
	OutputPaths      []string `yaml:"output_paths"`
	ErrorOutputPaths []string `yaml:"error_output_paths"`
}

func Load(path string) (Config, error) {
	if strings.TrimSpace(path) == "" {
		return Config{}, errors.New("config path is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode config %q: %w", path, err)
	}

	applyDefaults(&cfg)
	if err := validate(cfg); err != nil {
		return Config{}, fmt.Errorf("validate config %q: %w", path, err)
	}

	return cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.App.Name == "" {
		cfg.App.Name = "multi-platform-ads-analytics"
	}
	if cfg.App.Environment == "" {
		cfg.App.Environment = "development"
	}
	if cfg.Server.Address == "" {
		cfg.Server.Address = ":8080"
	}
	if cfg.Server.ReadHeaderTimeoutSeconds <= 0 {
		cfg.Server.ReadHeaderTimeoutSeconds = 10
	}
	if cfg.Server.ShutdownTimeoutSeconds <= 0 {
		cfg.Server.ShutdownTimeoutSeconds = 10
	}
	if cfg.Meta.GraphBaseURL == "" {
		cfg.Meta.GraphBaseURL = "https://graph.facebook.com"
	}
	if cfg.Logging.Level == "" {
		cfg.Logging.Level = "info"
	}
	if cfg.Logging.Encoding == "" {
		cfg.Logging.Encoding = "json"
	}
	if len(cfg.Logging.OutputPaths) == 0 {
		cfg.Logging.OutputPaths = []string{"stdout"}
	}
	if len(cfg.Logging.ErrorOutputPaths) == 0 {
		cfg.Logging.ErrorOutputPaths = []string{"stderr"}
	}
}

func validate(cfg Config) error {
	if strings.TrimSpace(cfg.Server.Address) == "" {
		return errors.New("server.address is required")
	}
	if strings.TrimSpace(cfg.Meta.GraphVersion) == "" {
		return errors.New("meta.graph_version is required")
	}

	switch cfg.Logging.Encoding {
	case "json", "console":
	default:
		return fmt.Errorf("logging.encoding must be json or console, got %q", cfg.Logging.Encoding)
	}

	return nil
}
