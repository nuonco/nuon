package metrics

import (
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"go.temporal.io/sdk/workflow"
)

func (w *writer) Incr(ctx workflow.Context, name string, value int, tags ...string) {
	workflow.SideEffect(ctx, func(workflow.Context) interface{} {
		w.MetricsWriter.Incr(name, value, metrics.ToTags(w.Tags, tags...))
		return nil
	})
}

func (w *writer) Decr(ctx workflow.Context, name string, value int, tags ...string) {
	workflow.SideEffect(ctx, func(workflow.Context) interface{} {
		w.MetricsWriter.Decr(name, value, metrics.ToTags(w.Tags, tags...))
		return true
	})
}

func (w *writer) Timing(ctx workflow.Context, name string, dur time.Duration, tags ...string) {
	workflow.SideEffect(ctx, func(workflow.Context) interface{} {
		w.MetricsWriter.Timing(name, dur, metrics.ToTags(w.Tags, tags...))
		return true
	})
}

func (w *writer) Event(ctx workflow.Context, e *statsd.Event) {
	workflow.SideEffect(ctx, func(workflow.Context) interface{} {
		w.MetricsWriter.Event(e)
		return true
	})
}

func (w *writer) Flush(ctx workflow.Context) {
	workflow.SideEffect(ctx, func(workflow.Context) interface{} {
		w.MetricsWriter.Flush()
		return true
	})
}
