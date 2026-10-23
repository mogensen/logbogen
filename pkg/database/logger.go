package database

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// slowQuery is how long a query may take before it is logged as a warning
const slowQuery = 200 * time.Millisecond

// gormLogger sends the gorm logs through slog, so database logs are structured like the rest. Queries are
// logged at debug level, slow ones as warnings and failed ones as errors, where a missing record is no error.
type gormLogger struct {
	slow time.Duration
}

// NewGormLogger creates the gorm logger that writes to the default slog logger.
func NewGormLogger() gormlogger.Interface {
	return &gormLogger{slow: slowQuery}
}

func (l *gormLogger) LogMode(gormlogger.LogLevel) gormlogger.Interface { return l }

func (l *gormLogger) Info(ctx context.Context, msg string, args ...any) {
	slog.InfoContext(ctx, "database", "message", msg, "args", args)
}

func (l *gormLogger) Warn(ctx context.Context, msg string, args ...any) {
	slog.WarnContext(ctx, "database", "message", msg, "args", args)
}

func (l *gormLogger) Error(ctx context.Context, msg string, args ...any) {
	slog.ErrorContext(ctx, "database", "message", msg, "args", args)
}

func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	failed := err != nil && !errors.Is(err, gorm.ErrRecordNotFound)
	if !failed && elapsed < l.slow && !slog.Default().Enabled(ctx, slog.LevelDebug) {
		return
	}

	sql, rows := fc()
	attrs := []any{"sql", sql, "rows", rows, "latency_ms", float64(elapsed.Microseconds()) / 1000}
	switch {
	case failed:
		slog.ErrorContext(ctx, "query failed", append(attrs, "error", err)...)
	case elapsed >= l.slow:
		slog.WarnContext(ctx, "slow query", attrs...)
	default:
		slog.DebugContext(ctx, "query", attrs...)
	}
}
