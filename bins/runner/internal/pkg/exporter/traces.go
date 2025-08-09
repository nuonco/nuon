package exporter

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/consumererror"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
)

func createTracesExporter(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Traces, error) {
	oce, err := newExporter(cfg, set)
	if err != nil {
		return nil, err
	}
	oCfg := cfg.(*Config)

	return exporterhelper.NewTraces(ctx, set, cfg,
		oce.pushTraces,
		exporterhelper.WithStart(oce.start),
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),

		// explicitly disable since we rely on http.Client timeout logic.
		exporterhelper.WithTimeout(exporterhelper.TimeoutConfig{Timeout: 0}),
		exporterhelper.WithRetry(oCfg.RetryConfig),
		exporterhelper.WithQueue(oCfg.QueueConfig))
}

func (b *baseExporter) pushTraces(ctx context.Context, td ptrace.Traces) error {
	tr := ptraceotlp.NewExportRequestFromTraces(td)

	req, err := tr.MarshalJSON()
	if err != nil {
		return consumererror.NewPermanent(err)
	}

	if err := b.apiClient.WriteOTELTraces(ctx, req); err != nil {
		return err
	}

	return nil
}
