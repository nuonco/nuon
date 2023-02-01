package cmd

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/go-common/config"
	"github.com/powertoolsdev/go-common/temporalzap"
	shared "github.com/powertoolsdev/workers-executors/internal"
	"github.com/powertoolsdev/workers-executors/internal/workflows/execute"
	"github.com/powertoolsdev/workers-executors/internal/workflows/plan"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"
)

var executorsCmd = &cobra.Command{
	Use:   "executors",
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

	l.Info("starting executor workers", zap.Any("config", cfg))
	if err := runExecutorWorkers(c, l, cfg, worker.InterruptCh()); err != nil {
		l.Error("error running worker", zap.Error(err))
	}
}

func runExecutorWorkers(c client.Client, log *zap.Logger, cfg shared.Config, interruptCh <-chan interface{}) error {
	otlpExporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:4317", cfg.HostIP)))
	if err != nil {
		return fmt.Errorf("unable to create otlptrace exporter: %w", err)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(otlpExporter),
	)

	otel.SetTracerProvider(tracerProvider)
	tracer := tracerProvider.Tracer("nuon.workers-executors")

	traceIntercepter, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{
		Tracer: tracer,
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

	planWkflow := plan.NewWorkflow(cfg)
	w.RegisterWorkflow(planWkflow.CreatePlan)
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
