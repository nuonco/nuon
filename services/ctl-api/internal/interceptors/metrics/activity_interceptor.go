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
	status := "ok"
	startTS := time.Now()
	tags := map[string]string{
		"status":        status,
		"activity":      info.ActivityType.Name,
		"namespace":     info.WorkflowNamespace,
		"workflow_type": info.WorkflowType.Name,
	}

	// NOTE(jm): we emit from a defer, so we can catch any type of panic and still emit metrics.
	defer func() {
		rec := recover()
		if rec != nil {
			tags["status"] = "panic"
		}

		a.mw.Incr("temporal_activity.status", metrics.ToTags(tags))
		a.mw.Timing("temporal_activity.latency", time.Since(startTS), metrics.ToTags(tags))

		if rec != nil {
			panic(rec)
		}
	}()

	resp, err := a.Next.ExecuteActivity(ctx, in)
	if err != nil {
		status = "error"
	}

	return resp, nil
}
