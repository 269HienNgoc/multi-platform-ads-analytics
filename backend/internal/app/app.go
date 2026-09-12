// Package app is the composition root for the API process.
package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/adapter/httpapi"
	postgresadapter "github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/adapter/postgres"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/application/health"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/config"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/logging"
	"go.uber.org/zap"
)

// Run loads configuration, builds dependencies, and owns the API lifecycle.
func Run(ctx context.Context, configFile string) int {
	cfg, err := config.Load(configFile)
	if err != nil {
		return logBootstrapError(err)
	}

	logger, err := logging.New(cfg.Log)
	if err != nil {
		return logBootstrapError(err)
	}

	exitCode := 0
	if err := run(ctx, cfg, logger); err != nil {
		logger.Error("application stopped with error", zap.Error(err))
		exitCode = 1
	}
	if err := logging.Sync(logger); err != nil {
		exitCode = 1
	}

	return exitCode
}

func run(ctx context.Context, cfg config.Config, logger *zap.Logger) (resultErr error) {
	database, err := postgresadapter.Open(ctx, cfg.Database, logger)
	if err != nil {
		return err
	}
	defer func() {
		resultErr = errors.Join(resultErr, database.Close())
	}()

	healthService := health.New(database)
	server := httpapi.NewServer(cfg.Server, healthService, logger)
	serverErrors := make(chan error, 1)

	go func() {
		logger.Info("http server started", zap.String("address", cfg.Server.Address))
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("application shutdown started")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutting down http server: %w", err)
		}
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serving http: %w", err)
		}
	}

	logger.Info("application shutdown completed")

	return nil
}

func logBootstrapError(err error) int {
	logger, loggerErr := zap.NewProduction()
	if loggerErr != nil {
		_, _ = fmt.Fprintln(os.Stderr, "application bootstrap failed")

		return 1
	}

	logger.Error("application bootstrap failed", zap.Error(err))
	_ = logging.Sync(logger)

	return 1
}
