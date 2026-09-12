// Package config centralizes application configuration loading and validation.
package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

const envPrefix = "ADS"

// Config contains all runtime configuration for the API process.
type Config struct {
	App      App      `mapstructure:"app"`
	Server   Server   `mapstructure:"server"`
	Database Database `mapstructure:"database"`
	Log      Log      `mapstructure:"log"`
}

// App contains service identity settings.
type App struct {
	Name        string `mapstructure:"name"`
	Environment string `mapstructure:"environment"`
}

// Server contains HTTP server limits and timeouts.
type Server struct {
	Address           string        `mapstructure:"address"`
	ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	WriteTimeout      time.Duration `mapstructure:"write_timeout"`
	IdleTimeout       time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout   time.Duration `mapstructure:"shutdown_timeout"`
	MaxHeaderBytes    int           `mapstructure:"max_header_bytes"`
}

// Database contains PostgreSQL and connection pool settings.
type Database struct {
	Host                  string        `mapstructure:"host"`
	Port                  int           `mapstructure:"port"`
	Name                  string        `mapstructure:"name"`
	User                  string        `mapstructure:"user"`
	Password              string        `mapstructure:"password"`
	SSLMode               string        `mapstructure:"ssl_mode"`
	Timezone              string        `mapstructure:"timezone"`
	ConnectTimeout        time.Duration `mapstructure:"connect_timeout"`
	MaxOpenConnections    int           `mapstructure:"max_open_connections"`
	MaxIdleConnections    int           `mapstructure:"max_idle_connections"`
	ConnectionMaxLifetime time.Duration `mapstructure:"connection_max_lifetime"`
	ConnectionMaxIdleTime time.Duration `mapstructure:"connection_max_idle_time"`
	SlowQueryThreshold    time.Duration `mapstructure:"slow_query_threshold"`
}

// Log contains Zap output settings.
type Log struct {
	Level    string `mapstructure:"level"`
	Encoding string `mapstructure:"encoding"`
}

// Load reads optional .env values, YAML configuration, and ADS_ environment overrides.
func Load(configFile string) (Config, error) {
	if err := loadDotEnv(); err != nil {
		return Config{}, err
	}

	v := viper.New()
	setDefaults(v)
	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return Config{}, fmt.Errorf("reading configuration: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("decoding configuration: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("validating configuration: %w", err)
	}

	return cfg, nil
}

// Validate checks runtime invariants before any external resource is created.
func (c Config) Validate() error {
	if strings.TrimSpace(c.App.Name) == "" {
		return errors.New("app name is required")
	}
	isSupportedEnvironment := c.App.Environment == "development" || c.App.Environment == "test" || c.App.Environment == "production"
	if !isSupportedEnvironment {
		return fmt.Errorf("unsupported app environment %q", c.App.Environment)
	}
	if err := c.Server.validate(); err != nil {
		return err
	}
	if err := c.Database.validate(); err != nil {
		return err
	}
	if err := c.Log.validate(); err != nil {
		return err
	}

	return nil
}

// DSN builds a PostgreSQL URL without exposing it to logs.
func (d Database) DSN() string {
	connectionURL := &url.URL{
		Scheme: "postgresql",
		User:   url.UserPassword(d.User, d.Password),
		Host:   net.JoinHostPort(d.Host, strconv.Itoa(d.Port)),
		Path:   d.Name,
	}

	query := connectionURL.Query()
	query.Set("sslmode", d.SSLMode)
	query.Set("timezone", d.Timezone)
	connectionURL.RawQuery = query.Encode()

	return connectionURL.String()
}

func loadDotEnv() error {
	_, err := os.Stat(".env")
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("checking .env: %w", err)
	}
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("loading .env: %w", err)
	}

	return nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "multi-platform-ads-analytics")
	v.SetDefault("app.environment", "development")
	v.SetDefault("server.address", "127.0.0.1:8080")
	v.SetDefault("server.read_header_timeout", "5s")
	v.SetDefault("server.read_timeout", "15s")
	v.SetDefault("server.write_timeout", "30s")
	v.SetDefault("server.idle_timeout", "60s")
	v.SetDefault("server.shutdown_timeout", "15s")
	v.SetDefault("server.max_header_bytes", 1<<20)
	v.SetDefault("database.host", "127.0.0.1")
	v.SetDefault("database.port", 5432)
	v.SetDefault("database.name", "multi_platform_ads")
	v.SetDefault("database.user", "postgres")
	v.SetDefault("database.password", "")
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.timezone", "UTC")
	v.SetDefault("database.connect_timeout", "10s")
	v.SetDefault("database.max_open_connections", 25)
	v.SetDefault("database.max_idle_connections", 10)
	v.SetDefault("database.connection_max_lifetime", "30m")
	v.SetDefault("database.connection_max_idle_time", "5m")
	v.SetDefault("database.slow_query_threshold", "500ms")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.encoding", "json")
}

func (s Server) validate() error {
	if strings.TrimSpace(s.Address) == "" {
		return errors.New("server address is required")
	}
	if s.ReadHeaderTimeout <= 0 || s.ReadTimeout <= 0 || s.WriteTimeout <= 0 || s.IdleTimeout <= 0 {
		return errors.New("server timeouts must be positive")
	}
	if s.ShutdownTimeout <= 0 {
		return errors.New("server shutdown timeout must be positive")
	}
	if s.MaxHeaderBytes <= 0 {
		return errors.New("server max header bytes must be positive")
	}

	return nil
}

func (d Database) validate() error {
	if strings.TrimSpace(d.Host) == "" || strings.TrimSpace(d.Name) == "" || strings.TrimSpace(d.User) == "" {
		return errors.New("database host, name, and user are required")
	}
	if d.Port < 1 || d.Port > 65535 {
		return fmt.Errorf("database port %d is outside the valid range", d.Port)
	}
	isSupportedSSLMode := d.SSLMode == "disable" || d.SSLMode == "require" || d.SSLMode == "verify-ca" || d.SSLMode == "verify-full"
	if !isSupportedSSLMode {
		return fmt.Errorf("unsupported database ssl mode %q", d.SSLMode)
	}
	if d.ConnectTimeout <= 0 || d.ConnectionMaxLifetime <= 0 || d.ConnectionMaxIdleTime <= 0 {
		return errors.New("database timeouts must be positive")
	}
	if d.MaxOpenConnections <= 0 {
		return errors.New("database max open connections must be positive")
	}
	if d.MaxIdleConnections < 0 || d.MaxIdleConnections > d.MaxOpenConnections {
		return errors.New("database max idle connections must be between zero and max open connections")
	}
	if d.SlowQueryThreshold <= 0 {
		return errors.New("database slow query threshold must be positive")
	}

	return nil
}

func (l Log) validate() error {
	isSupportedLevel := l.Level == "debug" || l.Level == "info" || l.Level == "warn" || l.Level == "error"
	if !isSupportedLevel {
		return fmt.Errorf("unsupported log level %q", l.Level)
	}
	if l.Encoding != "json" && l.Encoding != "console" {
		return fmt.Errorf("unsupported log encoding %q", l.Encoding)
	}

	return nil
}
