package log

import (
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/telemetry"
)

func New(cfg *internal.Config, lc fx.Lifecycle, telemetryConfig *telemetry.Config) (*zap.Logger, error) {
	var (
		l   *zap.Logger
		err error
	)

	l, err = zap.NewProduction()
	if cfg.LogLevel == "DEBUG" {
		l, err = zap.NewDevelopment()
	}
	if err != nil {
		return nil, fmt.Errorf("unable to initialize logger: %w", err)
	}
	core, err := telemetry.NewLifecycleLogCore(lc, telemetryConfig)
	if err != nil {
		return nil, err
	}
	l = l.WithOptions(zap.WrapCore(func(existing zapcore.Core) zapcore.Core {
		return zapcore.NewTee(existing, core)
	}))
	zap.ReplaceGlobals(l)
	return l, nil
}
