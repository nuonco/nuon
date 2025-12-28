package testworker

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/example"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/handler"
)

func (e *EnqueueTestSuite) TestCancelSignalBeforeExecution() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())

	queue, err := e.service.Client.Create(ctx, &client.CreateQueueRequest{
		OwnerID:     generics.GetFakeObj[string](),
		OwnerType:   generics.GetFakeObj[string](),
		Namespace:   defaultNamespace,
		MaxInFlight: 1,
		MaxDepth:    100,
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), queue)
	require.Nil(e.T(), e.service.Client.QueueReady(ctx, queue.ID))

	blockingSignal, err := e.service.Client.EnqueueSignal(ctx, queue.ID, &example.ControllableSignal{
		ShouldBlock: true,
	})
	require.Nil(e.T(), err)

	targetSignal, err := e.service.Client.EnqueueSignal(ctx, queue.ID, &example.ControllableSignal{
		ShouldBlock: false,
	})
	require.Nil(e.T(), err)

	cancelResp, err := e.service.Client.CancelSignal(ctx, targetSignal.ID)
	require.Nil(e.T(), err)
	require.NotNil(e.T(), cancelResp)

	err = e.service.Client.CompleteSignal(ctx, blockingSignal.ID, example.ControllableSignalUpdateName)
	require.Nil(e.T(), err)
}

func (e *EnqueueTestSuite) TestCancelSignalMidFlight() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())

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

	targetSignal, err := e.service.Client.EnqueueSignal(ctx, queue.ID, &example.ControllableSignal{
		ShouldBlock: true,
	})
	require.Nil(e.T(), err)

	cancelResp, err := e.service.Client.CancelSignal(ctx, targetSignal.ID)
	require.Nil(e.T(), err)
	require.NotNil(e.T(), cancelResp)

	finishedResp, err := e.service.Client.AwaitSignal(ctx, targetSignal.ID)
	assert.Nil(e.T(), finishedResp)
	assert.NotNil(e.T(), err)
	assert.Contains(e.T(), err.Error(), "execute method failed")
}

func (e *EnqueueTestSuite) TestCancelSignalAfterCompletion() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())

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

	targetSignal, err := e.service.Client.EnqueueSignal(ctx, queue.ID, &example.ExampleSignal{
		Arg1: generics.GetFakeObj[string](),
		Arg2: generics.GetFakeObj[string](),
	})
	require.Nil(e.T(), err)

	finishedResp, err := e.service.Client.AwaitSignal(ctx, targetSignal.ID)
	require.Nil(e.T(), err)
	require.NotNil(e.T(), finishedResp)

	cancelResp, err := e.service.Client.CancelSignal(ctx, targetSignal.ID)
	assert.Nil(e.T(), cancelResp)
	assert.NotNil(e.T(), err)
	assert.Contains(e.T(), err.Error(), handler.ErrAlreadyExecuted.Error())
}
