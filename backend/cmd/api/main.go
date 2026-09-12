package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/config"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/httpapi"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/logging"
	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/workflow"
	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "./config/config.yaml", "path to YAML configuration file")
	flag.Parse()

	bootstrap, _ := zap.NewProduction()
	defer func() { _ = bootstrap.Sync() }()

	cfg, err := config.Load(*configPath)
	if err != nil {
		bootstrap.Error("failed to load configuration", zap.String("config_path", *configPath), zap.Error(err))
		os.Exit(1)
	}

	logger, err := logging.New(cfg.Logging)
	if err != nil {
		bootstrap.Error("failed to initialize zap logger", zap.Error(err))
		os.Exit(1)
	}
	defer func() { _ = logger.Sync() }()

	service := workflow.NewService()
	api := httpapi.NewServer(service, logger)

	httpServer := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           api.Handler(),
		ReadHeaderTimeout: time.Duration(cfg.Server.ReadHeaderTimeoutSeconds) * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("campaign automation API started",
			zap.String("app", cfg.App.Name),
			zap.String("environment", cfg.App.Environment),
			zap.String("address", cfg.Server.Address),
		)
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server stopped unexpectedly", zap.Error(err))
			return
		}
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeoutSeconds)*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", zap.Error(err))
		return
	}

	logger.Info("campaign automation API stopped")
}
