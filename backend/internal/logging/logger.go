// Package logging creates the application's structured Zap logger.
package logging

import (
	"errors"
	"fmt"
	"syscall"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a logger from validated application configuration.
func New(cfg config.Log) (*zap.Logger, error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("parsing log level: %w", err)
	}

	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(level)
	zapConfig.Encoding = cfg.Encoding
	zapConfig.OutputPaths = []string{"stdout"}
	zapConfig.ErrorOutputPaths = []string{"stderr"}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("building zap logger: %w", err)
	}

	return logger, nil
}

// Sync flushes buffered logs and ignores unsupported terminal sync operations.
func Sync(logger *zap.Logger) error {
	err := logger.Sync()
	if errors.Is(err, syscall.EINVAL) || errors.Is(err, syscall.ENOTTY) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("syncing zap logger: %w", err)
	}

	return nil
}
