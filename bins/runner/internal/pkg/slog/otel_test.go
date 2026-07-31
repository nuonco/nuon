package slog

import (
	"context"
	"testing"

	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/sdk/log"

	runnerlog "github.com/nuonco/nuon/pkg/runner/log"
)

type recordingProcessor struct {
	emitted int
	flushed int
	stopped int
}

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
	processor := &auditProcessor{next: next}
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
