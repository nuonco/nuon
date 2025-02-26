package validate

import (
	"context"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/metrics"
)

var _ interceptor.ActivityInboundInterceptor = (*actInterceptor)(nil)

type actInterceptor struct {
	interceptor.ActivityInboundInterceptorBase

	mw metrics.Writer
	v  *validator.Validate
	l  *zap.Logger
}

func (a *actInterceptor) Init(outbound interceptor.ActivityOutboundInterceptor) error {
	return a.Next.Init(outbound)
}

func (a *actInterceptor) ExecuteActivity(
	ctx context.Context,
	in *interceptor.ExecuteActivityInput,
) (interface{}, error) {
	if len(in.Args) < 1 {
		return a.Next.ExecuteActivity(ctx, in)
	}

	argType := reflect.TypeOf(in.Args[0])
	switch argType.Kind() {
	case reflect.Struct, reflect.Pointer:
		info := activity.GetInfo(ctx)
		tags := map[string]string{
			"status":        "ok",
			"activity":      info.ActivityType.Name,
			"namespace":     info.WorkflowNamespace,
			"workflow_type": info.WorkflowType.Name,
		}
		defer a.mw.Incr("temporal_activity.request_validation", metrics.ToTags(tags))

		if err := a.v.StructCtx(ctx, in.Args[0]); err != nil {
			tags["status"] = "error"
			a.l.Error("invalid activity request",
				zap.Error(err),
				zap.String("activity", info.ActivityType.Name),
				zap.String("namespace", info.WorkflowNamespace),
				zap.String("workflow_type", info.WorkflowType.Name),
				zap.Any("request_object", in.Args[0]),
			)

			// TODO(jm): eventually, this will return once we know there are _zero_ instances of bad
			// requests in production. However, doing that now means we might be relying on some
			// unvalidateatble requests in prod.
			//
			return nil, temporal.NewNonRetryableApplicationError(
				"validate error",
				"validate error",
				errors.Wrap(err, "unable to validate activity in middleware"))
		}
	default:
	}

	return a.Next.ExecuteActivity(ctx, in)
}
