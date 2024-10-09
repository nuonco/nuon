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

	defaultVersion string = "v1"
)

// DefaultLogger is used to return a default configured logger with standard attributes that can be used in all contexts
//
// NOTE(jm): this is not an FX provider, but should be used at call sites that need specific loggers.
func DefaultLogger(lp *log.LoggerProvider, settings *settings.Settings, cfg *internal.Config, typ LoggerType) *slog.Logger {
	logger := otelslog.NewLogger(string(typ),
		otelslog.WithVersion(defaultVersion),
		otelslog.WithLoggerProvider(lp),
	)

	return logger
}
