// Package postgres implements PostgreSQL infrastructure through GORM.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/config"
	"go.uber.org/zap"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Client owns the GORM handle and its underlying connection pool.
type Client struct {
	db *gorm.DB
}

// Open connects to PostgreSQL, configures the pool, and verifies connectivity.
func Open(ctx context.Context, cfg config.Database, logger *zap.Logger) (*Client, error) {
	db, err := gorm.Open(
		gormpostgres.Open(cfg.DSN()),
		&gorm.Config{Logger: newGORMLogger(logger, cfg.SlowQueryThreshold)},
	)
	if err != nil {
		return nil, fmt.Errorf("opening postgres connection: %w", err)
	}

	pool, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("accessing postgres pool: %w", err)
	}
	pool.SetMaxOpenConns(cfg.MaxOpenConnections)
	pool.SetMaxIdleConns(cfg.MaxIdleConnections)
	pool.SetConnMaxLifetime(cfg.ConnectionMaxLifetime)
	pool.SetConnMaxIdleTime(cfg.ConnectionMaxIdleTime)

	connectCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()
	if err := pool.PingContext(connectCtx); err != nil {
		closeErr := pool.Close()
		if closeErr != nil {
			return nil, errors.Join(
				fmt.Errorf("pinging postgres: %w", err),
				fmt.Errorf("closing failed postgres connection: %w", closeErr),
			)
		}

		return nil, fmt.Errorf("pinging postgres: %w", err)
	}

	return &Client{db: db}, nil
}

// Ping verifies that PostgreSQL can accept a request using the caller's context.
func (c *Client) Ping(ctx context.Context) error {
	pool, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("accessing postgres pool: %w", err)
	}
	if err := pool.PingContext(ctx); err != nil {
		return fmt.Errorf("pinging postgres: %w", err)
	}

	return nil
}

// Close releases the PostgreSQL connection pool.
func (c *Client) Close() error {
	pool, err := c.db.DB()
	if err != nil {
		return fmt.Errorf("accessing postgres pool: %w", err)
	}
	if err := pool.Close(); err != nil {
		return fmt.Errorf("closing postgres pool: %w", err)
	}

	return nil
}
