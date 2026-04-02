package testworker

import (
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/example"
)

func (e *EnqueueTestSuite) TestEnqueueAndProcessNSignals() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())

	// create a queue
	queue, err := e.service.Client.Create(ctx, &client.CreateQueueRequest{
		OwnerID:     generics.GetFakeObj[string](),
		OwnerType:   generics.GetFakeObj[string](),
		Namespace:   defaultNamespace,
		MaxInFlight: 5,
		MaxDepth:    100,
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), queue)

	// wait for queue to be ready
	err = e.service.Client.QueueReady(ctx, queue.ID)
	require.Nil(e.T(), err)

	// enqueue N signals
	const numSignals = 10
	signalIDs := make([]string, 0, numSignals)

	for i := 0; i < numSignals; i++ {
		resp, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
			QueueID: queue.ID,
			Signal: &example.ExampleSignal{
				Arg1: generics.GetFakeObj[string](),
				Arg2: generics.GetFakeObj[string](),
			},
		})
		require.Nil(e.T(), err)
		require.NotNil(e.T(), resp)
		require.NotEmpty(e.T(), resp.ID)

		signalIDs = append(signalIDs, resp.ID)
	}

	// poll each signal until finished
	timeout := 5 * time.Second
	for _, id := range signalIDs {
		status, err := e.service.Client.PollSignal(ctx, id, &client.PollSignalOptions{
			Timeout:      &timeout,
			PollInterval: 500 * time.Millisecond,
		})
		require.Nil(e.T(), err)
		require.True(e.T(), status.Finished)
	}

	// verify DB status is success for all signals
	for _, id := range signalIDs {
		var qs app.QueueSignal
		res := e.service.DB.WithContext(ctx).First(&qs, "id = ?", id)
		require.Nil(e.T(), res.Error)
		require.Equal(e.T(), app.StatusSuccess, qs.Status.Status)
	}
}

func (e *EnqueueTestSuite) TestFailingSignalUpdatesDBStatus() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())

	// create a queue
	q, err := e.service.Client.Create(ctx, &client.CreateQueueRequest{
		OwnerID:     generics.GetFakeObj[string](),
		OwnerType:   generics.GetFakeObj[string](),
		Namespace:   defaultNamespace,
		MaxInFlight: 5,
		MaxDepth:    100,
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), q)

	err = e.service.Client.QueueReady(ctx, q.ID)
	require.Nil(e.T(), err)

	// enqueue a failing signal
	resp, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal: &example.FailingSignal{
			Reason: "test failure",
		},
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), resp)

	// poll until the signal finishes (the queue worker logs execute errors but
	// still marks the status in the DB)
	timeout := 5 * time.Second
	status, err := e.service.Client.PollSignal(ctx, resp.ID, &client.PollSignalOptions{
		Timeout:      &timeout,
		PollInterval: 500 * time.Millisecond,
	})
	require.Nil(e.T(), err)
	require.True(e.T(), status.Finished)

	// verify the DB has the error status persisted
	var qs app.QueueSignal
	res := e.service.DB.WithContext(ctx).First(&qs, "id = ?", resp.ID)
	require.Nil(e.T(), res.Error)
	require.Equal(e.T(), app.StatusError, qs.Status.Status)
}
