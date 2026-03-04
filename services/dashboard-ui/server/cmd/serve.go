package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/fxmodules"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the dashboard BFF server",
	RunE: func(cmd *cobra.Command, args []string) error {
		app := fx.New(
			fxmodules.InfrastructureModule,
			fxmodules.MiddlewaresModule,
			fxmodules.ServicesModule,
			fxmodules.APIModule,
		)
		app.Run()
		return nil
	},
}
