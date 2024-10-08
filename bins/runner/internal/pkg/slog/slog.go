package slog

import (
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/sdk/log"

	"github.com/powertoolsdev/mono/bins/runner/internal"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"
)

type LoggerType string

const (
	LoggerTypeSystem LoggerType = "system"
	LoggerTypeJob    LoggerType = "job"
)

// DefaultLogger is used to return a default configured logger with standard attributes that can be used in all contexts
//
// NOTE(jm): this is not an FX provider, but should be used at call sites that need specific loggers.
func DefaultLogger(lp *log.LoggerProvider, settings *settings.Settings, cfg *internal.Config, typ LoggerType) *slog.Logger {
	logger := otelslog.NewLogger(string(typ),
		otelslog.WithLoggerProvider(lp))

	logger = logger.With(slog.String("log_type", string(typ)))
	logger = logger.With(slog.String("runner_id", cfg.RunnerID))
	logger = logger.With(slog.String("runner_version", cfg.GitRef))

	for k, v := range settings.Metadata {
		logger = logger.With(slog.String(k, v))
	}

	return logger
}
