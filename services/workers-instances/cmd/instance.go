package cmd

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/common/temporalzap"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/sender"
	shared "github.com/powertoolsdev/mono/services/workers-instances/internal"
	"github.com/powertoolsdev/mono/services/workers-instances/internal/provision"
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
	var cfg shared.Config

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

func runInstanceWorkers(c client.Client, cfg shared.Config, interruptCh <-chan interface{}) error {
	w := worker.New(c, "instance", worker.Options{
		MaxConcurrentActivityExecutionSize: 1,
	})

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid instance config: %w", err)
	}

	var (
		n   sender.NotificationSender
		err error
	)

	l := zap.L()

	switch cfg.Env {
	case config.Local, config.Development:
		l.Info("using noop notification sender")
		n = sender.NewNoopSender()
	default:
		n, err = sender.NewSlackSender(cfg.DeploymentBotsSlackWebhookURL, l)
		if err != nil {
			l.Warn("failed to create slack notifier, using noop", zap.Error(err))
			n = sender.NewNoopSender()
		}
	}

	v := validator.New()
	wkflow := provision.NewWorkflow(cfg)
	w.RegisterWorkflow(wkflow.Provision)
	w.RegisterActivity(provision.NewActivities(v, n))

	if err := w.Run(interruptCh); err != nil {
		return err
	}
	return nil
}
