package testworker

import (
	"context"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/example"
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

func (e *EnqueueTestSuite) TestPauseAndResume() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())

	// create the queue
	q, err := e.service.Client.Create(ctx, &client.CreateQueueRequest{
		OwnerID:     generics.GetFakeObj[string](),
		OwnerType:   generics.GetFakeObj[string](),
		Namespace:   defaultNamespace,
		MaxInFlight: 5,
		MaxDepth:    100,
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), q)

	// wait for the event loop to be ready
	err = e.service.Client.QueueReady(ctx, q.ID)
	require.Nil(e.T(), err)

	// Pause
	err = e.service.Client.Pause(ctx, q.ID)
	require.Nil(e.T(), err)

	// Check DB
	var dbQ app.Queue
	err = e.service.DB.Where("id = ?", q.ID).First(&dbQ).Error
	require.Nil(e.T(), err)
	require.True(e.T(), dbQ.Paused)

	// Resume
	err = e.service.Client.Resume(ctx, q.ID)
	require.Nil(e.T(), err)

	// Check DB
	err = e.service.DB.Where("id = ?", q.ID).First(&dbQ).Error
	require.Nil(e.T(), err)
	require.False(e.T(), dbQ.Paused)
}

func (e *EnqueueTestSuite) TestPausePreventsProcessing() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())

	// create the queue
	q, err := e.service.Client.Create(ctx, &client.CreateQueueRequest{
		OwnerID:     generics.GetFakeObj[string](),
		OwnerType:   generics.GetFakeObj[string](),
		Namespace:   defaultNamespace,
		MaxInFlight: 5,
		MaxDepth:    100,
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), q)

	// wait for the event loop to be ready
	err = e.service.Client.QueueReady(ctx, q.ID)
	require.Nil(e.T(), err)

	// Pause
	err = e.service.Client.Pause(ctx, q.ID)
	require.Nil(e.T(), err)

	// Enqueue signal
	enqueueResp, err := e.service.Client.EnqueueSignal(ctx, q.ID, &example.ExampleSignal{
		Arg1: generics.GetFakeObj[string](),
		Arg2: generics.GetFakeObj[string](),
	})
	require.Nil(e.T(), err)

	// Try to await signal with a short timeout. It should TIMEOUT because queue is paused.
	shortCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	_, err = e.service.Client.AwaitSignal(shortCtx, enqueueResp.ID)
	require.Error(e.T(), err) // Expect timeout

	// Resume
	err = e.service.Client.Resume(ctx, q.ID)
	require.Nil(e.T(), err)

	// Await signal (should succeed now)
	finishedResp, err := e.service.Client.AwaitSignal(ctx, enqueueResp.ID)
	require.Nil(e.T(), err)
	require.NotNil(e.T(), finishedResp)
}
