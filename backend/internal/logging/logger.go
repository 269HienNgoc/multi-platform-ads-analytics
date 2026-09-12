package logging

import (
	"fmt"

	"github.com/269HienNgoc/multi-platform-ads-analytics/backend/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(cfg config.LoggingConfig) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if err := level.Set(cfg.Level); err != nil {
		return nil, fmt.Errorf("invalid logging level %q: %w", cfg.Level, err)
	}

	var zapCfg zap.Config
	if cfg.Development {
		zapCfg = zap.NewDevelopmentConfig()
	} else {
		zapCfg = zap.NewProductionConfig()
	}

	zapCfg.Level = zap.NewAtomicLevelAt(level)
	zapCfg.Encoding = cfg.Encoding
	zapCfg.OutputPaths = cfg.OutputPaths
	zapCfg.ErrorOutputPaths = cfg.ErrorOutputPaths

	logger, err := zapCfg.Build()
	if err != nil {
		return nil, fmt.Errorf("build zap logger: %w", err)
	}

	return logger, nil
}
