package cmd

import (
	"fmt"

	"github.com/powertoolsdev/go-common/config"
	"github.com/powertoolsdev/go-common/temporalzap"
	shared "github.com/powertoolsdev/workers-executors/internal"
	"github.com/powertoolsdev/workers-executors/internal/workflows/execute"
	"github.com/powertoolsdev/workers-executors/internal/workflows/plan"
	"github.com/spf13/cobra"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
)

var executorsCmd = &cobra.Command{
	Use:   "deployment",
	Short: "Run the executor workers",
	Run:   deploymentRun,
}

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(executorsCmd)
}

func deploymentRun(cmd *cobra.Command, args []string) {
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

	l.Info("starting deployment workers", zap.Any("config", cfg))
	if err := runExecutorWorkers(c, cfg, worker.InterruptCh()); err != nil {
		l.Error("error running worker", zap.Error(err))
	}
}

func runExecutorWorkers(c client.Client, cfg shared.Config, interruptCh <-chan interface{}) error {
	w := worker.New(c, "deployment", worker.Options{
		MaxConcurrentActivityExecutionSize: 1,
	})

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid deployment config: %w", err)
	}

	planWkflow := plan.NewWorkflow(cfg)
	w.RegisterWorkflow(planWkflow.Plan)
	planActs := plan.NewActivities()
	w.RegisterActivity(planActs)

	executeWkflow := execute.NewWorkflow(cfg)
	w.RegisterWorkflow(executeWkflow.ExecutePlan)
	executeActs := execute.NewActivities()
	w.RegisterActivity(executeActs)

	if err := w.Run(interruptCh); err != nil {
		return err
	}

	return nil
}
