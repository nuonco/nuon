package metrics

import (
	"context"
	"errors"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/interceptor"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

var _ interceptor.ActivityInboundInterceptor = (*actInterceptor)(nil)

type actInterceptor struct {
	interceptor.ActivityInboundInterceptorBase

	mw metrics.Writer
	l  *zap.Logger
}

func (a *actInterceptor) Init(outbound interceptor.ActivityOutboundInterceptor) error {
	return a.Next.Init(outbound)
}

func (a *actInterceptor) ExecuteActivity(
	ctx context.Context,
	in *interceptor.ExecuteActivityInput,
) (interface{}, error) {
	info := activity.GetInfo(ctx)
	status := "ok"
	startTS := time.Now()
	tags := map[string]string{
		"status":        status,
		"activity":      info.ActivityType.Name,
		"namespace":     info.WorkflowNamespace,
		"workflow_type": info.WorkflowType.Name,
	}

	// Inject a MetricContext so the GORM metrics plugin can emit per-query metrics
	// from worker activities (same key used by HTTP middleware).
	metricCtx := &cctx.MetricContext{
		Endpoint:   info.ActivityType.Name,
		Method:     "temporal",
		RequestURI: info.WorkflowType.Name,
		Context:    "worker",
		SignalType: info.ActivityType.Name,
	}
	ctx = context.WithValue(ctx, keys.MetricsKey, metricCtx)

	// NOTE(jm): we emit from a defer, so we can catch any type of panic and still emit metrics.
	defer func() {
		rec := recover()
		if rec != nil {
			tags["status"] = "panic"
		}

		a.mw.Timing("temporal_activity.latency", time.Since(startTS), metrics.ToTags(tags))

		if rec != nil {
			panic(rec)
		}
	}()

	resp, err := a.Next.ExecuteActivity(ctx, in)
	if err != nil {
		tags["status"] = "error"

		if errors.Is(err, gorm.ErrRecordNotFound) {
			tags["not_found"] = "true"
		}

		var vErr validator.ValidationErrors
		if errors.As(err, &vErr) {
			tags["validation_error"] = "true"
		}

		return nil, err
	}

	return resp, err
}
