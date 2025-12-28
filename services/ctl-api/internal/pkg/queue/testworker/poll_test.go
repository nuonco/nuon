package testworker

import (
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/example"
)

func (e *EnqueueTestSuite) TestPollSignalSuccess() {
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

	go func() {
		time.Sleep(2 * time.Second)
		_ = e.service.Client.CompleteSignal(ctx, targetSignal.ID, example.ControllableSignalUpdateName)
	}()

	pollResp, err := e.service.Client.PollSignal(ctx, targetSignal.ID, &client.PollSignalOptions{
		PollInterval: 500 * time.Millisecond,
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), pollResp)
	assert.True(e.T(), pollResp.Finished)
}

func (e *EnqueueTestSuite) TestPollSignalWithTimeout() {
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

	timeout := 3 * time.Second
	pollResp, err := e.service.Client.PollSignal(ctx, targetSignal.ID, &client.PollSignalOptions{
		Timeout:      &timeout,
		PollInterval: 500 * time.Millisecond,
	})
	assert.Nil(e.T(), pollResp)
	assert.NotNil(e.T(), err)
	assert.ErrorIs(e.T(), err, client.ErrSignalTimeout)

	err = e.service.Client.CompleteSignal(ctx, targetSignal.ID, example.ControllableSignalUpdateName)
	require.Nil(e.T(), err)
}

func (e *EnqueueTestSuite) TestPollSignalImmediateCompletion() {
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

	timeout := 5 * time.Second
	pollResp, err := e.service.Client.PollSignal(ctx, targetSignal.ID, &client.PollSignalOptions{
		Timeout:      &timeout,
		PollInterval: 500 * time.Millisecond,
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), pollResp)
	assert.True(e.T(), pollResp.Finished)
	assert.False(e.T(), pollResp.Canceled)
}
