package cctx

import (
	"fmt"
)

const (
	MetricsKey string = "metrics"
)

var ErrMetricContextNotFound error = fmt.Errorf("metric context not found")

type MetricContext struct {
	Endpoint string
	Method   string
	OrgID    string
	RunnerID string
	Context  string

	DBQueryCount int
}

func MetricsContextFromGinContext(ctx ValueContext) (*MetricContext, error) {
	metrics := ctx.Value(MetricsKey)
	if metrics == nil {
		return nil, ErrMetricContextNotFound
	}

	return metrics.(*MetricContext), nil
}
