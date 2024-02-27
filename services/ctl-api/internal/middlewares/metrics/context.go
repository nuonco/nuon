package metrics

import (
	"context"
	"fmt"
)

const (
	ContextKey string = "metrics"
)

var ErrMetricContextNotFound error = fmt.Errorf("metric context not found")

type MetricContext struct {
	Endpoint string
	Method   string
	OrgID    string
	Context  string
}

func FromContext(ctx context.Context) (*MetricContext, error) {
	metrics := ctx.Value(ContextKey)
	if metrics == nil {
		return nil, ErrMetricContextNotFound

	}

	return metrics.(*MetricContext), nil
}
