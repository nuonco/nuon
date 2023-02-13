package cmd

import (
	"fmt"

	"github.com/powertoolsdev/go-common/temporalzap"
	"github.com/powertoolsdev/go-config/pkg/config"
	shared "github.com/powertoolsdev/workers-executors/internal"
	execwaypoint "github.com/powertoolsdev/workers-executors/internal/workflows/execute/waypoint"
	planwaypoint "github.com/powertoolsdev/workers-executors/internal/workflows/plan/waypoint"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
)

var executorsCmd = &cobra.Command{
	Use:    "executors",
	Short:  "Run the executor workers",
	Run:    executorsRun,
	PreRun: config.ConfigureService[*cobra.Command],
}

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(executorsCmd)
}

func executorsRun(cmd *cobra.Command, args []string) {
	l := zap.L()
	var cfg shared.Config

	if err := config.LoadInto(cmd.Flags(), &cfg); err != nil {
		l.Fatal("failed to load config", zap.Error(err))
	}

	if err := cfg.Validate(); err != nil {
		l.Fatal("failed to validate config", zap.Error(err))
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

	l.Info("starting executor workers", zap.Any("config", cfg))
	if err := runExecutorWorkers(c, l, cfg, worker.InterruptCh()); err != nil {
		l.Error("error running worker", zap.Error(err))
	}
}

func runExecutorWorkers(c client.Client, log *zap.Logger, cfg shared.Config, interruptCh <-chan interface{}) error {
	traceIntercepter, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{
		Tracer: otel.Tracer(""),
	})
	if err != nil {
		return fmt.Errorf("unable to get tracing interceptor: %w", err)
	}
	w := worker.New(c, "executors", worker.Options{
		MaxConcurrentActivityExecutionSize: 1,
		Interceptors:                       []interceptor.WorkerInterceptor{traceIntercepter},
	})

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid executors config: %w", err)
	}

	planWkflow := planwaypoint.NewWorkflow(cfg)
	w.RegisterWorkflow(planWkflow.CreatePlan)
	planActs := planwaypoint.NewActivities()
	w.RegisterActivity(planActs)

	executeWkflow := execwaypoint.NewWorkflow(cfg)
	w.RegisterWorkflow(executeWkflow.ExecutePlan)
	executeActs := execwaypoint.NewActivities()
	w.RegisterActivity(executeActs)

	if err := w.Run(interruptCh); err != nil {
		return err
	}

	return nil
}
