package postgres

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type gormLogger struct {
	logger        *zap.Logger
	level         gormlogger.LogLevel
	slowThreshold time.Duration
}

func newGORMLogger(logger *zap.Logger, slowThreshold time.Duration) gormlogger.Interface {
	return &gormLogger{
		logger:        logger.Named("gorm"),
		level:         gormlogger.Warn,
		slowThreshold: slowThreshold,
	}
}

func (l *gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	clone := *l
	clone.level = level

	return &clone
}

func (l *gormLogger) Info(_ context.Context, _ string, _ ...any) {}

func (l *gormLogger) Warn(_ context.Context, _ string, _ ...any) {}

func (l *gormLogger) Error(_ context.Context, _ string, _ ...any) {}

func (l *gormLogger) Trace(_ context.Context, started time.Time, query func() (string, int64), err error) {
	elapsed := time.Since(started)
	_, rows := query()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		if l.level >= gormlogger.Error {
			l.logger.Error(
				"database query failed",
				zap.Duration("duration", elapsed),
				zap.Int64("rows", rows),
				zap.Error(err),
			)
		}

		return
	}
	if elapsed < l.slowThreshold || l.level < gormlogger.Warn {
		return
	}

	l.logger.Warn(
		"slow database query",
		zap.Duration("duration", elapsed),
		zap.Int64("rows", rows),
	)
}
