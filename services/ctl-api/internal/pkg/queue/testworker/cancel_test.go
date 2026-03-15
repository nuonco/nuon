package testworker

import (
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/example"
)

func (e *EnqueueTestSuite) TestCancelSignal() {
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

	// Enqueue a blocking signal that will not complete on its own.
	resp, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal:  &example.ControllableSignal{ShouldBlock: true},
	})
	require.Nil(e.T(), err)

	// Cancel the signal. Cancellation is safe to call at any point — before or
	// during execution. The handler checks h.canceled on execute and, if blocking,
	// the executing context is canceled to unblock Execute.
	_, err = e.service.Client.CancelSignal(ctx, resp.ID)
	require.Nil(e.T(), err)

	// Poll until the handler marks itself finished (execute runs and returns the cancel error).
	timeout := 30 * time.Second
	status, err := e.service.Client.PollSignal(ctx, resp.ID, &client.PollSignalOptions{
		Timeout: &timeout,
	})
	require.Nil(e.T(), err)
	require.True(e.T(), status.Finished)
	require.True(e.T(), status.Canceled)
}
