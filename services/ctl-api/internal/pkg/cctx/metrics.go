package cctx

import (
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/keys"
)

var ErrMetricContextNotFound error = fmt.Errorf("metric context not found")

type MetricContext struct {
	Endpoint string
	Method   string
	OrgID    string
	RunnerID string
	Context  string

	DBQueryCount int
	IsPanic      bool
	IsTimeout    bool
	IsDeprecated bool
}

func MetricsContextFromGinContext(ctx ValueContext) (*MetricContext, error) {
	metrics := ctx.Value(keys.MetricsKey)
	if metrics == nil {
		return nil, ErrMetricContextNotFound
	}

	return metrics.(*MetricContext), nil
}
