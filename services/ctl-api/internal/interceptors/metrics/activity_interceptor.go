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

		a.mw.Timing("temporal_activity.latency", time.Since(startTS), metrics.ToTags(tags))

		if rec != nil {
			panic(rec)
		}
	}()

	resp, err := a.Next.ExecuteActivity(ctx, in)
	if err != nil {
		status = "error"

		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = "error_not_found"
		}
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			status = "error_duplicate_key"
		}

		var vErr validator.ValidationErrors
		if errors.As(err, &vErr) {
			status = "error_validation"
		}

		tags[status] = status
		return nil, err
	}

	return resp, err
}
