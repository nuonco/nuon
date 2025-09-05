package validate

import (
	"context"
	"reflect"

	"github.com/go-playground/validator/v10"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/interceptor"
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
		defer func() {
			a.mw.Incr("temporal_activity.request_validation", metrics.ToTags(tags))
		}()

		if err := a.v.StructCtx(ctx, in.Args[0]); err != nil {
			tags["status"] = "error"
			a.l.Error("invalid activity request",
				zap.Error(err),
				zap.String("activity", info.ActivityType.Name),
				zap.String("namespace", info.WorkflowNamespace),
				zap.String("workflow_type", info.WorkflowType.Name),
				zap.Any("request_object", in.Args[0]),
			)
			
			//NOTE(sk): we dont want to break with errors now, uncomment only after metrics is clear
			// return nil, temporal.NewNonRetryableApplicationError(
			// 	"invalid activity request",
			// 	"invalid activity request",
			// 	err,
			// )
		}
	default:
	}

	return a.Next.ExecuteActivity(ctx, in)
}
