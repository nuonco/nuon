package metrics

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/interceptor"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/metrics"
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

	startTS := time.Now()
	resp, err := a.Next.ExecuteActivity(ctx, in)

	status := "ok"
	if err != nil {
		status = "error"
	}

	tags := map[string]string{
		"status":        status,
		"activity":      info.ActivityType.Name,
		"namespace":     info.WorkflowNamespace,
		"workflow_type": info.WorkflowType.Name,
	}

	a.mw.Incr("temporal_activity.status", metrics.ToTags(tags))
	a.mw.Timing("temporal_activity.latency", time.Since(startTS), metrics.ToTags(tags))
	return resp, nil
}
