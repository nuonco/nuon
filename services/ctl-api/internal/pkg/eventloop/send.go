package eventloop

import (
	"context"

	"go.uber.org/zap"

	enumspb "go.temporal.io/api/enums/v1"

	"github.com/powertoolsdev/mono/pkg/metrics"
)

func (a *evClient) Send(ctx context.Context, id string, signal Signal) {
	if err := a.v.Struct(signal); err != nil {
		a.mw.Incr("event_loop.signal", metrics.ToStatusTag("invalid signal"))
		a.l.Error("invalid signal", zap.Error(err))
		return
	}

	if err := signal.PropagateContext(ctx); err != nil {
		a.mw.Incr("event_loop.signal", metrics.ToStatusTag("unable to propagate"))
		a.l.Error("unable to propagate", zap.Error(err))
		return
	}

	if signal.Restart() {
		status, err := a.client.GetWorkflowStatusInNamespace(ctx,
			signal.Namespace(),
			signal.WorkflowID(id),
			"",
		)
		if err != nil {
			a.mw.Incr("event_loop_signal", metrics.ToStatusTag("unable_to_get_workflow"))
		}

		if status != enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING {
			if err := a.startEventLoop(ctx, id, signal); err != nil {
				a.mw.Incr("event_loop_signal", metrics.ToStatusTag("unable_to_start_event_loop"))
			}
		}
	}

	if signal.Start() {
		if err := a.startEventLoop(ctx, id, signal); err != nil {
			a.mw.Incr("event_loop_signal", metrics.ToStatusTag("unable_to_start_event_loop"))
		}
	}

	err := a.client.SignalWorkflowInNamespace(ctx,
		signal.Namespace(),
		signal.WorkflowID(id),
		"",
		id,
		signal,
	)
	if err != nil {
		a.mw.Incr("event_loop_signal", metrics.ToStatusTag("unable_to_send"))
	}
}
