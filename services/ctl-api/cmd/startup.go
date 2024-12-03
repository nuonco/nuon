package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
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
	start := time.Now()
	l := zap.L()

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

	// NOTE(fd): in prod and stage, we want the job container to persist for at least 60s to ensure
	// the datadog agent picks up on its existence. We don't want this job to take longer than necessary
	// though so we calculate it's runtime so we only sleep for as long as necessary to reach the 60s threshold.
	if os.Getenv("ENV") == "prod" || os.Getenv("ENV") == "stage" {
		minRunLen := time.Duration(time.Second * 60)
		runTime := time.Now().Sub(start)
		if runTime < minRunLen {
			sleepFor := minRunLen - runTime
			l.Info(fmt.Sprintf("sleeping for %d seconds to ensure data dog metrics are flushed", sleepFor))
			time.Sleep(sleepFor)
		}
	}
}
