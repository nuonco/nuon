package slog

import (
	"log/slog"

	"go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/bins/runner/internal"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"
)

type SystemParams struct {
	fx.In
}

func NewSystemProvider(params SystemParams) (*log.LoggerProvider, error) {
	lp := log.NewLoggerProvider()

	return lp, nil
}

type SystemLoggerParams struct {
	fx.In

	Settings *settings.Settings
	LP       *log.LoggerProvider `name:"system"`
	Cfg      *internal.Config
}

func NewSystemLogger(params SystemLoggerParams) *slog.Logger {
	return DefaultLogger(params.LP, params.Settings, params.Cfg, LoggerTypeSystem)
}
