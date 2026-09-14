package app

import (
	"context"
	"errors"
	"fmt"
	"os"

	postgresadapter "github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/adapter/postgres"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/config"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/logging"
	"go.uber.org/zap"
)

// RunMigration loads configuration and executes one schema migration action.
func RunMigration(ctx context.Context, configFile, migrationsDirectory, action string, steps int) int {
	cfg, err := config.Load(configFile)
	if err != nil {
		return logBootstrapError(err)
	}

	logger, err := logging.New(cfg.Log)
	if err != nil {
		return logBootstrapError(err)
	}

	database, err := postgresadapter.Open(ctx, cfg.Database, logger)
	if err != nil {
		logger.Error("database migration connection failed", zap.Error(err))
		_ = logging.Sync(logger)

		return 1
	}
	version, dirty, migrationErr := database.RunMigration(
		ctx,
		migrationsDirectory,
		postgresadapter.MigrationAction(action),
		steps,
	)
	closeErr := database.Close()
	if migrationErr != nil || closeErr != nil {
		migrationErr = errors.Join(migrationErr, closeErr)
		logger.Error("database migration failed", zap.Error(migrationErr))
		_ = logging.Sync(logger)

		return 1
	}

	logger.Info(
		"database migration completed",
		zap.String("action", action),
		zap.Uint("version", version),
		zap.Bool("dirty", dirty),
	)
	if err := logging.Sync(logger); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "syncing migration logger:", err)

		return 1
	}

	return 0
}
