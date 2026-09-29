package cmd

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/ctl-api/internal/fxmodules"
)

type cli struct{}

func (c *cli) providers() []fx.Option {
	return []fx.Option{
		fxmodules.InfrastructureModule,
		fxmodules.HelpersModule,
	}
}
