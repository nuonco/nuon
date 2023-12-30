package config

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// flagger represents anything that can return a pointer to a pflag FlagSet
// typically, this would be a *cobra.Command
type flagger interface {
	Flags() *pflag.FlagSet
}

const configureServiceErrTemplate = `{"level":"error","ts":%d,"msg":"failed to setup service", "error": "%s"}\n`

func ConfigureService[T flagger](cmd T, args []string) {
	cfg, err := loadConfig(cmd)
	if err != nil {
		fmt.Printf(configureServiceErrTemplate, time.Now().Unix(), err)
		os.Exit(1)
	}

	l, err := configureLogger(cfg)
	if err != nil {
		fmt.Printf(configureServiceErrTemplate, time.Now().Unix(), err)
		os.Exit(1)
	}

	configureOtel(cfg, l)
}

func loadConfig(cmd flagger) (*Base, error) {
	var cfg Base

	if err := LoadInto(cmd.Flags(), &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &cfg, nil
}

func configureLogger(cfg *Base) (*zap.Logger, error) {
	var (
		l   *zap.Logger
		err error
	)

	switch cfg.Env {
	case Development:
		l, err = zap.NewDevelopment()
	default:
		zCfg := zap.NewProductionConfig()

		var lvl zapcore.Level
		lvl, err = zapcore.ParseLevel(cfg.LogLevel)
		if err == nil {
			// only set the level if it was set correctly on the config
			zCfg.Level.SetLevel(lvl)
		}

		l, err = zCfg.Build()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to instantiate logger: %w", err)
	}

	zap.ReplaceGlobals(l)

	return l, nil
}

func configureOtel(cfg *Base, l *zap.Logger) {
	otlpExporter, err := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:4317", cfg.TraceAddress)),
	)
	if err != nil {
		l.Fatal("unable to create otlptrace exporter", zap.Error(err))
	}

	resource, err :=
		resource.Merge(
			resource.Default(),
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(cfg.ServiceName),
				semconv.ServiceNamespaceKey.String(cfg.ServiceOwner),
				semconv.ServiceVersionKey.String(cfg.Version),
			),
		)
	if err != nil {
		l.Fatal("unable to create trace resource", zap.Error(err))
	}

	sampler := trace.ParentBased(trace.TraceIDRatioBased(cfg.TraceSampleRate))

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(otlpExporter, trace.WithMaxExportBatchSize(cfg.TraceMaxBatchCount)),
		trace.WithResource(resource),
		trace.WithSampler(sampler),
	)

	otel.SetTracerProvider(tracerProvider)
}
