package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

func (e *EnqueueTestSuite) TestPauseQueue() {
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

	err = e.service.Client.QueueReady(ctx, queue.ID)
	require.Nil(e.T(), err)

	status, err := e.service.Client.GetQueueStatus(ctx, queue.ID)
	require.Nil(e.T(), err)
	require.True(e.T(), status.Ready)
	require.False(e.T(), status.Paused)

	err = e.service.Client.Pause(ctx, queue.ID)
	require.Nil(e.T(), err)

	status, err = e.service.Client.GetQueueStatus(ctx, queue.ID)
	require.Nil(e.T(), err)
	require.True(e.T(), status.Paused)
}

func (e *EnqueueTestSuite) TestUnpauseQueue() {
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

	err = e.service.Client.QueueReady(ctx, queue.ID)
	require.Nil(e.T(), err)

	err = e.service.Client.Pause(ctx, queue.ID)
	require.Nil(e.T(), err)

	status, err := e.service.Client.GetQueueStatus(ctx, queue.ID)
	require.Nil(e.T(), err)
	require.True(e.T(), status.Paused)

	err = e.service.Client.Unpause(ctx, queue.ID)
	require.Nil(e.T(), err)

	status, err = e.service.Client.GetQueueStatus(ctx, queue.ID)
	require.Nil(e.T(), err)
	require.False(e.T(), status.Paused)
}

func (e *EnqueueTestSuite) TestPauseAndUnpauseQueue() {
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

	err = e.service.Client.QueueReady(ctx, queue.ID)
	require.Nil(e.T(), err)

	err = e.service.Client.Pause(ctx, queue.ID)
	require.Nil(e.T(), err)

	status, err := e.service.Client.GetQueueStatus(ctx, queue.ID)
	require.Nil(e.T(), err)
	require.True(e.T(), status.Paused)
	require.True(e.T(), status.Ready)
	require.False(e.T(), status.Stopped)

	err = e.service.Client.Unpause(ctx, queue.ID)
	require.Nil(e.T(), err)

	status, err = e.service.Client.GetQueueStatus(ctx, queue.ID)
	require.Nil(e.T(), err)
	require.False(e.T(), status.Paused)
	require.True(e.T(), status.Ready)
	require.False(e.T(), status.Stopped)
}
