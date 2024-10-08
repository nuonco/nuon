package analytics

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

// TemporalWriter wraps a ContextWriter, and calls it's methods via Temporal's workflow.SideEffect.
// It's methods accept workflow.Context instead of the standard context.Context.
type TemporalWriter struct {
	writer *ContextWriter
}

func NewTemporalWriter(writeKey string, l *zap.Logger) *TemporalWriter {
	return &TemporalWriter{
		writer: NewContextWriter(writeKey, l),
	}
}

func (cw *TemporalWriter) Identify(ctx workflow.Context) {
	workflow.SideEffect(ctx, func(workflow.Context) any {
		cw.writer.Identify(ctx)
		return nil
	})
}

func (cw *TemporalWriter) Group(ctx workflow.Context) {
	workflow.SideEffect(ctx, func(workflow.Context) any {
		cw.writer.Group(ctx)
		return nil
	})
}

func (cw *TemporalWriter) Track(ctx workflow.Context, event Event, properties map[string]any) {
	workflow.SideEffect(ctx, func(workflow.Context) any {
		cw.writer.Track(ctx, event, properties)
		return nil
	})
}
