package metrics

import (
	"fmt"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type Writer interface {
	// dogstatsd metrics
	Incr(workflow.Context, string, ...string)
	Decr(workflow.Context, string, ...string)
	Gauge(workflow.Context, string, float64, ...string)
	Timing(workflow.Context, string, time.Duration, ...string)

	// datadog specific
	Event(workflow.Context, *statsd.Event)

	Flush(workflow.Context)
}

type writer struct {
	v *validator.Validate

	MetricsWriter metrics.Writer
	Tags          map[string]string
}

var _ Writer = (*writer)(nil)

// New returns a workflow writer, which uses the underlying metrics writer to emit metrics
func New(v *validator.Validate, opts ...writerOption) (*writer, error) {
	l := zap.L()
	mw, err := metrics.New(v,
		metrics.WithDisable(false),
		metrics.WithLogger(l),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create default metrics writer: %w", err)
	}

	r := &writer{
		v:             v,
		MetricsWriter: mw,
	}
	for idx, opt := range opts {
		if err := opt(r); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	if err := r.v.Struct(r); err != nil {
		return nil, fmt.Errorf("unable to validate writer: %w", err)
	}

	return r, nil
}

type writerOption func(*writer) error

func WithTags(tags map[string]string) writerOption {
	return func(w *writer) error {
		w.Tags = tags
		return nil
	}
}

func WithMetricsWriter(mw metrics.Writer) writerOption {
	return func(w *writer) error {
		w.MetricsWriter = mw
		return nil
	}
}
