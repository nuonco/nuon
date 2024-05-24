package cmd

import (
	"context"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/db"
)

func (c *cli) registerStartup() error {
	runStartupCmd := &cobra.Command{
		Use:   "startup",
		Short: "startup hook that is run on deploy",
		Run:   c.runStartup,
	}
	rootCmd.AddCommand(runStartupCmd)
	return nil
}

func (c *cli) runStartup(cmd *cobra.Command, _ []string) {
	l := zap.L()
	l.Info("disabling view usage to run migrations")
	db.DisableViews()

	// for now, run the automigrate script
	providers := []fx.Option{
		fx.Provide(db.NewAutoMigrate),
		fx.Invoke(func(l *zap.Logger, db *db.AutoMigrate, shutdowner fx.Shutdowner) {
			ctx := context.Background()
			ctx, cancelFn := context.WithTimeout(ctx, time.Minute*5)
			defer cancelFn()

			code := 0
			if err := db.Execute(ctx); err != nil {
				l.Error("unable to auto migrate", zap.Error(err))
				code = 1
			}

			if err := shutdowner.Shutdown(fx.ExitCode(code)); err != nil {
				l.Error("unable to shut down", zap.Error(err))
			}
		}),
	}
	providers = append(providers, c.providers()...)
	fx.New(providers...).Run()

	if os.Getenv("ENV") == "prod" || os.Getenv("ENV") == "stage" {
		l.Info("sleeping for 1 minute to ensure data dog metrics are flushed")
		time.Sleep(time.Minute)
	}
}
