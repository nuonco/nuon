package cctx

import (
	"fmt"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/pkg/cctx/keys"
)

type MetricContext struct {
	Endpoint   string
	Method     string
	RequestURI string
	OrgID      string
	IsPanic    bool
	IsTimeout  bool
}

func MetricsContextFromGinContext(ctx ValueContext) (*MetricContext, error) {
	v := ctx.Value(keys.MetricsKey)
	if v == nil {
		return nil, fmt.Errorf("metrics context not found")
	}
	return v.(*MetricContext), nil
}
