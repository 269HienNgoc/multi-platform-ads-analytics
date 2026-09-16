// Package app is the composition root for the API process.
package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/adapter/httpapi"
	metaadapter "github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/adapter/meta"
	postgresadapter "github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/adapter/postgres"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/application/automation"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/application/catalog"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/application/health"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/application/providersync"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/config"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/domain/ads"
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
	catalogStore := postgresadapter.NewCatalogStore(database)
	catalogService := catalog.New(catalogStore)
	automationStore := postgresadapter.NewAutomationStore(database)
	automationService := automation.New(automationStore)
	workerCtx, stopWorker := context.WithCancel(ctx)
	defer stopWorker()
	var metaSyncUseCases httpapi.ProviderSyncService
	workerDone := make(chan struct{})
	close(workerDone)
	if cfg.Meta.Enabled {
		metaClient, connectorErr := metaadapter.New(metaadapter.Config{
			BaseURL: cfg.Meta.BaseURL, Version: cfg.Meta.Version,
			AccessToken: cfg.Meta.AccessToken, Timeout: cfg.Meta.Timeout,
		})
		if connectorErr != nil {
			return fmt.Errorf("creating meta connector: %w", connectorErr)
		}
		providerSyncStore := postgresadapter.NewProviderSyncStore(database)
		metaSyncService := providersync.New(ads.PlatformMeta, metaClient, providerSyncStore)
		metaSyncUseCases = metaSyncService
		workerDone = make(chan struct{})
		go runProviderSyncWorker(workerCtx, cfg.Meta.SyncInterval, metaSyncService, logger, workerDone)
	}
	server := httpapi.NewServer(
		cfg.Server,
		healthService,
		catalogService,
		automationService,
		metaSyncUseCases,
		logger,
	)
	serverErrors := make(chan error, 1)

	go func() {
		logger.Info("http server started", zap.String("address", cfg.Server.Address))
		serverErrors <- server.ListenAndServe()
	}()

	var lifecycleErr error
	select {
	case <-ctx.Done():
		logger.Info("application shutdown started")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			lifecycleErr = fmt.Errorf("shutting down http server: %w", err)
		}
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			lifecycleErr = fmt.Errorf("serving http: %w", err)
		}
	}

	stopWorker()
	<-workerDone
	if lifecycleErr != nil {
		return lifecycleErr
	}
	logger.Info("application shutdown completed")

	return nil
}

func runProviderSyncWorker(
	ctx context.Context,
	interval time.Duration,
	service *providersync.Service,
	logger *zap.Logger,
	done chan<- struct{},
) {
	defer close(done)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	synchronize := func() {
		run, result, err := service.Sync(ctx)
		if err != nil {
			logger.Error("scheduled meta sync failed", zap.Error(err))

			return
		}
		logger.Info(
			"scheduled meta sync completed",
			zap.String("run_id", run.ID),
			zap.Int("accounts", result.Accounts),
			zap.Int("campaigns", result.Campaigns),
			zap.Int("warnings", len(result.Warnings)),
		)
	}

	synchronize()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			synchronize()
		}
	}
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
