package temporal

import (
	"context"
	"time"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

const (
	defaultSignalSendTimeout time.Duration = time.Second * 5
)

func (e *evClient) Send(wCtx workflow.Context, id string, signal eventloop.Signal) {
	workflow.SideEffect(wCtx, func(workflow.Context) interface{} {
		ctx := context.Background()
		ctx, cancelFn := context.WithTimeout(ctx, defaultSignalSendTimeout)
		defer cancelFn()

		ctx = cctx.ContextFromWorkflowContext(ctx, wCtx)
		e.evClient.Send(ctx, id, signal)
		return nil
	})
}
