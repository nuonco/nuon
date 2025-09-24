package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/client"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/example"
)

func (e *EnqueueTestSuite) TestEnqueueSignalOK() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())

	// create a queue and wait for it to be ready
	queue, err := e.service.Client.Create(ctx, &client.CreateQueueRequest{
		OwnerID:     generics.GetFakeObj[string](),
		OwnerType:   generics.GetFakeObj[string](),
		Namespace:   defaultNamespace,
		MaxInFlight: 5,
		MaxDepth:    100,
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), queue)
	require.Nil(e.T(), e.service.Client.QueueReady(ctx, queue.ID))

	// enqueue a signal
	enqueueResp, err := e.service.Client.EnqueueSignal(ctx, queue.ID, &example.ExampleSignal{
		Arg1: generics.GetFakeObj[string](),
		Arg2: generics.GetFakeObj[string](),
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), enqueueResp)

	// now wait for the queue signal
	finishedResp, err := e.service.Client.AwaitSignal(ctx, enqueueResp.ID)
	require.Nil(e.T(), err)
	require.NotNil(e.T(), finishedResp)
}
