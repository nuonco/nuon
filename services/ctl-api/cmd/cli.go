package cmd

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/ctl-api/internal/fxmodules"
)

type cli struct{}

// providers returns the base set of fx modules needed for all commands.
// This includes infrastructure (db, temporal, logging) and domain helpers.
func (c *cli) providers() []fx.Option {
	return []fx.Option{
		fxmodules.InfrastructureModule,
		fxmodules.HelpersModule,
	}
}
