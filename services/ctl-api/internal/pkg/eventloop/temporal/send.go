package temporal

import (
	"context"
	"time"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

const (
	defaultSignalSendTimeout time.Duration = time.Second * 5
)

func (e *evClient) Send(wCtx workflow.Context, id string, signal eventloop.Signal) {
	workflow.SideEffect(wCtx, func(workflow.Context) interface{} {
		// ensure the child context has our values
		acctID, err := cctx.AccountIDFromWorkflowContext(wCtx)
		if err != nil {
			e.l.Error("no account id found", zap.Error(err))
		}

		orgID, err := cctx.OrgIDFromWorkflowContext(wCtx)
		if err != nil {
			e.l.Error("no org id found", zap.Error(err))
		}

		ctx := context.Background()
		ctx, cancelFn := context.WithTimeout(ctx, defaultSignalSendTimeout)
		defer cancelFn()

		ctx = cctx.SetAccountIDContext(ctx, acctID)
		ctx = cctx.SetOrgIDContext(ctx, orgID)

		e.evClient.Send(ctx, id, signal)
		return nil
	})
}
