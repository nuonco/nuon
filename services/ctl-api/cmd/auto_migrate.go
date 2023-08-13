package cmd

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/db"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

func (c *cli) registerAutoMigrate() error {
	var autoMigrateCmd = &cobra.Command{
		Use:   "auto-migrate",
		Short: "auto-migrate using GORM",
		Run:   c.autoMigrate,
	}
	rootCmd.AddCommand(autoMigrateCmd)
	return nil
}

func (c *cli) autoMigrate(cmd *cobra.Command, _ []string) {
	providers := []fx.Option{
		fx.Provide(db.NewAutoMigrate),
		fx.Invoke(func(*db.AutoMigrate) {}),
	}
	providers = append(providers, c.providers()...)
	fx.New(providers...).Run()
}
