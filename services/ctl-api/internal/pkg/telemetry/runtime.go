package telemetry

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/shirou/gopsutil/v4/process"
	otelruntime "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

func StartRuntimeMetrics(cfg *Config, provider metric.MeterProvider) error {
	if cfg.Endpoint == "" {
		return nil
	}
	if err := otelruntime.Start(otelruntime.WithMeterProvider(provider)); err != nil {
		return err
	}
	p := &process.Process{Pid: int32(os.Getpid())}
	created, err := p.CreateTime()
	if err != nil {
		otel.Handle(fmt.Errorf("read process start time: %w", err))
		return nil
	}
	observed := time.Now()
	uptimeAtObservation := observed.Sub(time.UnixMilli(created))
	_, err = provider.Meter("github.com/nuonco/nuon/services/ctl-api/process").Float64ObservableGauge(
		"process.uptime",
		metric.WithUnit("s"),
		metric.WithDescription("The time the process has been running."),
		metric.WithFloat64Callback(func(_ context.Context, observer metric.Float64Observer) error {
			observer.Observe(max(0, (uptimeAtObservation + time.Since(observed)).Seconds()))
			return nil
		}),
	)
	return err
}
