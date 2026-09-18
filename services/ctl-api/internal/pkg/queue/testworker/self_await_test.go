package testworker

import (
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/example"
)

func (e *EnqueueTestSuite) TestSelfAwaitGuard() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())

	queue, err := e.service.Client.Create(ctx, &client.CreateQueueRequest{
		OwnerID:     generics.GetFakeObj[string](),
		OwnerType:   "components",
		Namespace:   defaultNamespace,
		MaxInFlight: 1,
		MaxDepth:    10,
	})
	require.NoError(e.T(), err)
	require.NoError(e.T(), e.queueReady(ctx, queue.ID))

	callerCtx := cctx.SetQueueIDContext(ctx, queue.ID)
	sig := &example.ExampleSignal{Arg1: "one", Arg2: "two"}
	_, err = e.service.Client.EnqueueSignal(callerCtx, &client.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  sig,
		Callback: callback.Ref{
			WorkflowID: "parent-workflow",
			SignalName: "child-finished",
			Namespace:  defaultNamespace,
		},
	})
	require.ErrorContains(e.T(), err, "self-await")
	var appErr *temporal.ApplicationError
	require.ErrorAs(e.T(), err, &appErr)
	require.True(e.T(), appErr.NonRetryable())

	resp, err := e.service.Client.EnqueueSignal(callerCtx, &client.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  sig,
	})
	require.NoError(e.T(), err)
	e.waitForSignalStatus(ctx, resp.ID, app.StatusSuccess)
}
