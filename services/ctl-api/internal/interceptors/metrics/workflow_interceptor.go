package metrics

import (
	"errors"
	"time"

	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/pkg/errs"
	"github.com/powertoolsdev/mono/pkg/metrics"
)

var _ interceptor.WorkflowInboundInterceptor = (*wfInterceptor)(nil)

type wfInterceptor struct {
	interceptor.WorkflowInboundInterceptorBase

	mw metrics.Writer
	l  *zap.Logger
}

func (a *wfInterceptor) Init(outbound interceptor.WorkflowOutboundInterceptor) error {
	return a.Next.Init(outbound)
}

func (a *wfInterceptor) ExecuteWorkflow(
	ctx workflow.Context,
	in *interceptor.ExecuteWorkflowInput,
) (interface{}, error) {
	info := workflow.GetInfo(ctx)
	startTS := time.Now()
	tags := map[string]string{
		"status":        "ok",
		"task_queue":    info.TaskQueueName,
		"namespace":     info.Namespace,
		"workflow_type": info.WorkflowType.Name,
	}
	status := "ok"

	// NOTE(jm): we emit from a defer, so we can catch any type of panic and still emit metrics.
	defer func() {
		rec := recover()
		tags["status"] = status
		if rec != nil {
			tags["status"] = "panic"
		}

		a.mw.Incr("temporal_workflow.status", metrics.ToTags(tags))
		a.mw.Timing("temporal_workflow.latency", time.Since(startTS), metrics.ToTags(tags))

		if rec != nil {
			// TODO(sdboyer) replace this with something that goes to dd
			if err, is := rec.(error); is {
				errs.ReportToSentry(err, &errs.SentryErrOptions{
					Tags: tags,
				})
			}
			panic(rec)
		}
	}()

	resp, err := a.Next.ExecuteWorkflow(ctx, in)
	if err != nil {
		status = "error"

		var vErr validator.ValidationErrors
		if errors.As(err, &vErr) {
			status = "error_validation"
		}
	}

	return resp, err
}
