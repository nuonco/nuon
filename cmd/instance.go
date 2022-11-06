package cmd

import (
	"fmt"

	"github.com/powertoolsdev/go-common/config"
	"github.com/powertoolsdev/go-common/temporalzap"
	"github.com/powertoolsdev/workers-instances/internal/provision"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
)

var instanceCmd = &cobra.Command{
	Use:   "instance",
	Short: "Run the instance workers",
	Run:   instanceRun,
}

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(instanceCmd)
}

func instanceRun(cmd *cobra.Command, args []string) {
	var cfg Config

	if err := config.LoadInto(cmd.Flags(), &cfg); err != nil {
		panic(fmt.Sprintf("failed to load config: %s", err))
	}

	var (
		l   *zap.Logger
		err error
	)
	switch cfg.Base.Env {
	case config.Local, config.Development:
		l, err = zap.NewDevelopment()
	default:
		l, err = zap.NewProduction()
	}
	zap.ReplaceGlobals(l)

	if err != nil {
		fmt.Printf("failed to instantiate logger: %v\n", err)
	}

	c, err := client.Dial(client.Options{
		HostPort:  cfg.TemporalHost,
		Namespace: cfg.TemporalNamespace,
		Logger:    temporalzap.NewLogger(l),
	})
	if err != nil {
		l.Fatal("failed to instantiate temporal client", zap.Error(err))
	}
	defer c.Close()

	l.Debug("starting instance workers", zap.Any("config", cfg))
	if err := runInstanceWorkers(c, cfg, worker.InterruptCh()); err != nil {
		l.Error("error running worker", zap.Error(err))
	}
}

func runInstanceWorkers(c client.Client, cfg Config, interruptCh <-chan interface{}) error {
	w := worker.New(c, "instance", worker.Options{
		MaxConcurrentActivityExecutionSize: 1,
	})

	if err := cfg.Cfg.Validate(); err != nil {
		return fmt.Errorf("invalid instance config: %w", err)
	}

	wkflow := provision.NewWorkflow(cfg.Cfg)
	w.RegisterWorkflow(wkflow.Provision)
	w.RegisterActivity(provision.NewActivities())

	if err := w.Run(interruptCh); err != nil {
		return err
	}
	return nil
}
