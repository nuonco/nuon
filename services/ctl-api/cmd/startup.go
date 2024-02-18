package cmd

import (
	"os"
	"time"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/db"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func (c *cli) registerStartup() error {
	var runStartupCmd = &cobra.Command{
		Use:   "startup",
		Short: "startup hook that is run on deploy",
		Run:   c.runStartup,
	}
	rootCmd.AddCommand(runStartupCmd)
	return nil
}

func (c *cli) runStartup(cmd *cobra.Command, _ []string) {
	// for now, run the automigrate script
	providers := []fx.Option{
		fx.Provide(db.NewAutoMigrate),
		fx.Invoke(func(*db.AutoMigrate) {}),
	}
	providers = append(providers, c.providers()...)
	fx.New(providers...).Run()

	if os.Getenv("ENV") == "prod" || os.Getenv("ENV") == "stage" {
		l := zap.L()
		l.Info("sleeping for 1 minute to ensure data dog metrics are flushed")
		time.Sleep(time.Minute)
	}
}
