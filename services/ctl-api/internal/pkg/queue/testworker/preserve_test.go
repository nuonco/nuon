package testworker

import (
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/example"
)

// TestStopAndRestartPreservesQueuedSignals verifies that a graceful queue stop
// preserves in-flight signals in the database and that a subsequent restart picks
// them back up and processes them to completion.
func (e *EnqueueTestSuite) TestStopAndRestartPreservesQueuedSignals() {
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

	// Enqueue a blocking signal — it will not complete until we send `complete`.
	resp, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  &example.ControllableSignal{ShouldBlock: true},
	})
	require.Nil(e.T(), err)

	// Gracefully stop the queue. The signal remains in the DB with status=queued
	// and its handler workflow continues running independently.
	err = e.service.Client.Stop(ctx, queue.ID)
	require.Nil(e.T(), err)

	// Restart the queue. requeueSignals will pick up the still-queued signal and
	// ensure its handler is running.
	err = e.service.Client.Restart(ctx, queue.ID)
	require.Nil(e.T(), err)

	err = e.service.Client.QueueReady(ctx, queue.ID)
	require.Nil(e.T(), err)

	// Unblock the signal by sending the `complete` update to its handler workflow.
	err = e.service.Client.CompleteSignal(ctx, resp.ID, example.ControllableSignalUpdateName)
	require.Nil(e.T(), err)

	timeout := 30 * time.Second
	status, err := e.service.Client.PollSignal(ctx, resp.ID, &client.PollSignalOptions{
		Timeout: &timeout,
	})
	require.Nil(e.T(), err)
	require.True(e.T(), status.Finished)
	require.False(e.T(), status.Canceled)
}

// TestTerminateAndRestartPreservesQueuedSignals verifies that a hard Temporal
// workflow termination does not lose queued signals. On restart, requeueSignals
// fetches them from the DB, restarts their handler workflows (which were also
// terminated), and processes them to completion.
func (e *EnqueueTestSuite) TestTerminateAndRestartPreservesQueuedSignals() {
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

	// Enqueue a blocking signal.
	resp, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  &example.ControllableSignal{ShouldBlock: true},
	})
	require.Nil(e.T(), err)

	// Hard-terminate the queue workflow. This also terminates child handler workflows.
	// The signal record in the DB retains status=queued.
	err = e.service.Client.Terminate(ctx, queue.ID, "test: hard termination")
	require.Nil(e.T(), err)

	// Restart the queue. Restart uses UpdateWithStart which creates a new workflow
	// instance. requeueSignals fetches the orphaned signal and calls StartHandlerIfNeeded
	// to recreate the terminated handler workflow.
	err = e.service.Client.Restart(ctx, queue.ID)
	require.Nil(e.T(), err)

	err = e.service.Client.QueueReady(ctx, queue.ID)
	require.Nil(e.T(), err)

	// Unblock the signal now that its handler workflow has been recreated.
	err = e.service.Client.CompleteSignal(ctx, resp.ID, example.ControllableSignalUpdateName)
	require.Nil(e.T(), err)

	timeout := 30 * time.Second
	status, err := e.service.Client.PollSignal(ctx, resp.ID, &client.PollSignalOptions{
		Timeout: &timeout,
	})
	require.Nil(e.T(), err)
	require.True(e.T(), status.Finished)
	require.False(e.T(), status.Canceled)
}
