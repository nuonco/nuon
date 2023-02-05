package cmd

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/go-common/config"
	"github.com/powertoolsdev/go-common/temporalzap"
	"github.com/powertoolsdev/go-sender"
	shared "github.com/powertoolsdev/workers-installs/internal"
	"github.com/powertoolsdev/workers-installs/internal/deprovision"
	"github.com/powertoolsdev/workers-installs/internal/provision"
	"github.com/powertoolsdev/workers-installs/internal/provision/runner"
	"github.com/powertoolsdev/workers-installs/internal/provision/sandbox"
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

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Run the install workers",
	Run:   installRun,
}

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(installCmd)
}

func installRun(cmd *cobra.Command, args []string) {
	var cfg shared.Config

	if err := config.LoadInto(cmd.Flags(), &cfg); err != nil {
		panic(fmt.Sprintf("failed to load config: %s", err))
	}

	var (
		l   *zap.Logger
		err error
	)
	switch cfg.Env {
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

	l.Debug("starting install workers", zap.Any("config", cfg))
	if err := runInstallWorkers(c, cfg, worker.InterruptCh()); err != nil {
		l.Error("error running worker", zap.Error(err))
	}
}

func runInstallWorkers(c client.Client, cfg shared.Config, interruptCh <-chan interface{}) error {
	otlpExporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:4317", cfg.HostIP)))
	if err != nil {
		return fmt.Errorf("unable to create otlptrace exporter: %w", err)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(otlpExporter),
	)

	otel.SetTracerProvider(tracerProvider)
	tracer := tracerProvider.Tracer("nuon.workers-installs")

	traceIntercepter, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{
		Tracer: tracer,
	})
	if err != nil {
		return fmt.Errorf("unable to get tracing interceptor: %w", err)
	}
	w := worker.New(c, "install", worker.Options{
		MaxConcurrentActivityExecutionSize: 1,
		Interceptors:                       []interceptor.WorkerInterceptor{traceIntercepter},
	})

	var (
		n sender.NotificationSender
	)

	l := zap.L()

	// NOTE(jdt): this isn't my favorite
	switch cfg.Env {
	case config.Local, config.Development:
		l.Info("using noop notification sender")
		n = sender.NewNoopSender()
	default:
		n, err = sender.NewSlackSender(cfg.InstallationBotsSlackWebhookURL, l)
		if err != nil {
			l.Warn("failed to create slack notifier, using noop", zap.Error(err))
			n = sender.NewNoopSender()
		}
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid install config: %w", err)
	}

	// register provision
	prWorkflow := provision.NewWorkflow(cfg)
	prRWorkflow := runner.NewWorkflow(cfg)
	prSWorkflow := sandbox.NewWorkflow(cfg)

	w.RegisterWorkflow(prWorkflow.Provision)
	w.RegisterWorkflow(prRWorkflow.ProvisionRunner)
	w.RegisterWorkflow(prSWorkflow.ProvisionSandbox)
	w.RegisterActivity(provision.NewProvisionActivities(cfg, n))
	w.RegisterActivity(sandbox.NewActivities(cfg))
	w.RegisterActivity(runner.NewActivities(cfg))

	// register deprovision
	dprWorkflow := deprovision.NewWorkflow(cfg)
	w.RegisterWorkflow(dprWorkflow.Deprovision)
	w.RegisterActivity(deprovision.NewActivities(n))

	if err := w.Run(interruptCh); err != nil {
		return err
	}
	return nil
}
