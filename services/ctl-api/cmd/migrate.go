package cmd

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/goose"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

func (c *cli) registerMigrate() error {
	var migrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "migrate",
		Run:   c.migrate,
	}
	rootCmd.AddCommand(migrateCmd)
	return nil
}

func (c *cli) migrate(cmd *cobra.Command, _ []string) {
	providers := []fx.Option{
		fx.Provide(goose.NewUp),
		fx.Invoke(func(*goose.Goose) {}),
	}
	providers = append(providers, c.providers()...)

	fx.New(providers...).Run()
}
