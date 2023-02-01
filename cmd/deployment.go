package cmd

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/go-common/config"
	"github.com/powertoolsdev/go-common/temporalzap"
	shared "github.com/powertoolsdev/workers-deployments/internal"
	start "github.com/powertoolsdev/workers-deployments/internal/start"
	"github.com/powertoolsdev/workers-deployments/internal/start/build"
	"github.com/powertoolsdev/workers-deployments/internal/start/instances"
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

var deploymentCmd = &cobra.Command{
	Use:   "deployment",
	Short: "Run the deployment workers",
	Run:   deploymentRun,
}

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(deploymentCmd)
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

	l.Debug("starting deployment workers", zap.Any("config", cfg))
	if err := runDeploymentWorkers(c, cfg, worker.InterruptCh()); err != nil {
		l.Error("error running worker", zap.Error(err))
	}
}

func runDeploymentWorkers(c client.Client, cfg shared.Config, interruptCh <-chan interface{}) error {
	otlpExporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:4317", cfg.NodeIP)))
	if err != nil {
		return fmt.Errorf("unable to create otlptrace exporter: %w", err)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(otlpExporter),
	)

	otel.SetTracerProvider(tracerProvider)
	tracer := tracerProvider.Tracer("nuon.workers-deployments")

	traceIntercepter, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{
		Tracer: tracer,
	})
	if err != nil {
		return fmt.Errorf("unable to get tracing interceptor: %w", err)
	}
	w := worker.New(c, "deployment", worker.Options{
		MaxConcurrentActivityExecutionSize: 1,
		Interceptors:                       []interceptor.WorkerInterceptor{traceIntercepter},
	})

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid deployment config: %w", err)
	}

	wkflow := start.NewWorkflow(cfg)
	w.RegisterWorkflow(wkflow.Start)
	w.RegisterActivity(start.NewActivities())

	bldWkflow := build.NewWorkflow(cfg)
	w.RegisterWorkflow(bldWkflow.Build)

	instancesWkflow := instances.NewWorkflow(cfg)
	w.RegisterWorkflow(instancesWkflow.ProvisionInstances)
	instancesActs := instances.NewActivities(cfg)
	w.RegisterActivity(instancesActs)

	if err := w.Run(interruptCh); err != nil {
		return err
	}

	return nil
}
