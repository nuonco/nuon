package fxmodules

import (
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

func newLogger(cfg *internal.Config) (*zap.Logger, error) {
	if cfg.LogLevel == "DEBUG" {
		return zap.NewDevelopment()
	}
	return zap.NewProduction()
}

var InfrastructureModule = fx.Module("infrastructure",
	fx.Provide(internal.NewConfig),
	fx.Provide(newLogger),
	fx.Provide(NewMetricsWriter),
)
