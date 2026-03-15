package testworker

import (
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/example"
)

func (e *EnqueueTestSuite) TestEnqueueSignal() {
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

	err = e.service.Client.QueueReady(ctx, queue.ID)
	require.Nil(e.T(), err)

	resp, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  &example.ExampleSignal{},
	})
	require.Nil(e.T(), err)
	require.NotEmpty(e.T(), resp.ID)
	require.NotEmpty(e.T(), resp.WorkflowID)
}

func (e *EnqueueTestSuite) TestEnqueueAndComplete() {
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

	err = e.service.Client.QueueReady(ctx, queue.ID)
	require.Nil(e.T(), err)

	resp, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  &example.ExampleSignal{},
	})
	require.Nil(e.T(), err)

	timeout := 30 * time.Second
	status, err := e.service.Client.PollSignal(ctx, resp.ID, &client.PollSignalOptions{
		Timeout: &timeout,
	})
	require.Nil(e.T(), err)
	require.True(e.T(), status.Finished)
	require.False(e.T(), status.Canceled)
}

func (e *EnqueueTestSuite) TestEnqueueOnStoppedQueue() {
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

	err = e.service.Client.QueueReady(ctx, queue.ID)
	require.Nil(e.T(), err)

	err = e.service.Client.Stop(ctx, queue.ID)
	require.Nil(e.T(), err)

	// Enqueue should auto-start the queue workflow via UpdateWithStart.
	resp, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  &example.ExampleSignal{},
	})
	require.Nil(e.T(), err)
	require.NotEmpty(e.T(), resp.ID)

	err = e.service.Client.QueueReady(ctx, queue.ID)
	require.Nil(e.T(), err)

	timeout := 30 * time.Second
	status, err := e.service.Client.PollSignal(ctx, resp.ID, &client.PollSignalOptions{
		Timeout: &timeout,
	})
	require.Nil(e.T(), err)
	require.True(e.T(), status.Finished)
}
