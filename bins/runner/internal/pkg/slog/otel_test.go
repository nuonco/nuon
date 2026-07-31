package slog

import (
	"context"
	"testing"
	"time"

	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/sdk/log"

	runnerlog "github.com/nuonco/nuon/pkg/runner/log"
)

type recordingProcessor struct {
	emitted int
	flushed int
	stopped int
}

type recordingExporter struct {
	exported int
}

func (e *recordingExporter) Export(_ context.Context, records []log.Record) error {
	e.exported += len(records)
	return nil
}

func (*recordingExporter) ForceFlush(context.Context) error { return nil }
func (*recordingExporter) Shutdown(context.Context) error   { return nil }

func (p *recordingProcessor) Enabled(context.Context, log.EnabledParameters) bool { return true }

func (p *recordingProcessor) OnEmit(context.Context, *log.Record) error {
	p.emitted++
	return nil
}

func (p *recordingProcessor) ForceFlush(context.Context) error {
	p.flushed++
	return nil
}

func (p *recordingProcessor) Shutdown(context.Context) error {
	p.stopped++
	return nil
}

func TestAuditProcessor(t *testing.T) {
	next := new(recordingProcessor)
	available := true
	processor := &auditProcessor{next: next, available: func() bool { return available }}
	ctx := context.Background()
	logger := log.NewLoggerProvider(log.WithProcessor(processor)).Logger("test")

	records := []otellog.Record{{}, {}, {}}
	records[1].AddAttributes(otellog.String(runnerlog.AuditAttr, "false"))
	records[2].AddAttributes(otellog.String(runnerlog.AuditAttr, runnerlog.AuditAttrValue))
	for i := range records {
		logger.Emit(ctx, records[i])
	}

	if next.emitted != 1 {
		t.Fatalf("forwarded %d records, want 1", next.emitted)
	}
	available = false
	logger.Emit(ctx, records[2])
	if next.emitted != 1 {
		t.Fatalf("forwarded %d records while unavailable, want 1", next.emitted)
	}
	if err := processor.ForceFlush(ctx); err != nil {
		t.Fatalf("ForceFlush() error = %v", err)
	}
	if err := processor.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if next.flushed != 1 || next.stopped != 1 {
		t.Fatalf("delegation counts = flush %d, shutdown %d; want 1 each", next.flushed, next.stopped)
	}
}

func TestAvailabilityExporterDropsQueuedRecordsAfterCollectorStops(t *testing.T) {
	available := true
	exporter := new(recordingExporter)
	gatedExporter := &availabilityExporter{Exporter: exporter, available: func() bool { return available }}
	batch := log.NewBatchProcessor(gatedExporter, log.WithExportInterval(time.Hour))
	processor := &auditProcessor{next: batch, available: func() bool { return available }}
	provider := log.NewLoggerProvider(log.WithProcessor(processor))
	logger := provider.Logger("test")

	var record otellog.Record
	record.AddAttributes(otellog.String(runnerlog.AuditAttr, runnerlog.AuditAttrValue))
	logger.Emit(context.Background(), record)
	available = false
	if err := provider.ForceFlush(context.Background()); err != nil {
		t.Fatalf("ForceFlush() error = %v", err)
	}
	if exporter.exported != 0 {
		t.Fatalf("exported %d queued records after collector stopped, want 0", exporter.exported)
	}
	if err := provider.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
}
