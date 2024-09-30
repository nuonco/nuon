package eventloop

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/metrics"
)

func (a *evClient) Send(ctx context.Context, id string, signal Signal) {
	if err := a.v.Struct(signal); err != nil {
		a.mw.Incr("event_loop.signal", metrics.ToStatusTag("invalid signal"))
		a.l.Error("invalid signal", zap.Error(err))
		return
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
		fmt.Printf("%+v\n", err)
		a.mw.Incr("event_loop_signal", metrics.ToStatusTag("unable_to_send"))
	}
}
