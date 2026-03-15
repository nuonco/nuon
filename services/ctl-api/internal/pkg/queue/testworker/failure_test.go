package testworker

import (
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/example"
)

// TestSignalFailure verifies that when a signal's Execute method returns an error
// the queue continues to process subsequent signals without crashing.
func (e *EnqueueTestSuite) TestSignalFailure() {
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

	// Enqueue a signal that returns an error from Execute.
	resp, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  &example.ControllableSignal{FailureMessage: "intentional failure"},
	})
	require.Nil(e.T(), err)

	// executeHandler defers h.finished = true even on error, so PollSignal
	// will eventually see Finished == true.
	timeout := 30 * time.Second
	status, err := e.service.Client.PollSignal(ctx, resp.ID, &client.PollSignalOptions{
		Timeout: &timeout,
	})
	require.Nil(e.T(), err)
	require.True(e.T(), status.Finished)

	// The queue worker logs the execute error and continues. Verify the queue is
	// still running by processing a follow-up signal.
	resp2, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  &example.ExampleSignal{},
	})
	require.Nil(e.T(), err)

	status2, err := e.service.Client.PollSignal(ctx, resp2.ID, &client.PollSignalOptions{
		Timeout: &timeout,
	})
	require.Nil(e.T(), err)
	require.True(e.T(), status2.Finished)
}

// TestSignalPanic verifies that when a signal's Execute method panics the queue
// worker handles the resulting workflow failure and continues processing subsequent signals.
func (e *EnqueueTestSuite) TestSignalPanic() {
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

	// Enqueue a signal that panics during Execute. With WorkflowPanicPolicy=FailWorkflow
	// the handler workflow fails. The queue worker logs the error and moves on.
	_, err = e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  &example.ControllableSignal{PanicMessage: "intentional panic"},
	})
	require.Nil(e.T(), err)

	// Give the panic time to propagate through the workflow.
	time.Sleep(3 * time.Second)

	// Verify the queue is still running by processing a follow-up signal.
	timeout := 30 * time.Second
	resp2, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  &example.ExampleSignal{},
	})
	require.Nil(e.T(), err)

	status2, err := e.service.Client.PollSignal(ctx, resp2.ID, &client.PollSignalOptions{
		Timeout: &timeout,
	})
	require.Nil(e.T(), err)
	require.True(e.T(), status2.Finished)
}

// TestSignalTimeout verifies that when a signal's Execute sleeps and returns an
// error the queue worker handles it and continues processing.
func (e *EnqueueTestSuite) TestSignalTimeout() {
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

	// Enqueue a signal that sleeps 2s then returns an error (simulates a slow/timed-out signal).
	resp, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  &example.ControllableSignal{Timeout: 2 * time.Second},
	})
	require.Nil(e.T(), err)

	// executeHandler defers h.finished = true, so PollSignal resolves after the sleep.
	timeout := 30 * time.Second
	status, err := e.service.Client.PollSignal(ctx, resp.ID, &client.PollSignalOptions{
		Timeout: &timeout,
	})
	require.Nil(e.T(), err)
	require.True(e.T(), status.Finished)

	// Verify the queue is still processing after the timeout signal.
	resp2, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  &example.ExampleSignal{},
	})
	require.Nil(e.T(), err)

	status2, err := e.service.Client.PollSignal(ctx, resp2.ID, &client.PollSignalOptions{
		Timeout: &timeout,
	})
	require.Nil(e.T(), err)
	require.True(e.T(), status2.Finished)
}
