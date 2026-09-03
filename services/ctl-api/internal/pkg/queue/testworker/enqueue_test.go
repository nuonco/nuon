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
	err = e.queueReady(ctx, queue.ID)
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

	for _, id := range signalIDs {
		e.waitForSignalStatus(ctx, id, app.StatusSuccess)
	}

	// verify DB status is success for all signals
	for _, id := range signalIDs {
		var qs app.QueueSignal
		res := e.service.DB.WithContext(ctx).First(&qs, "id = ?", id)
		require.Nil(e.T(), res.Error)
		require.Equal(e.T(), app.StatusSuccess, qs.Status.Status)
	}
}

func (e *EnqueueTestSuite) TestEnqueueSignalIdempotency() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())

	queue, err := e.service.Client.Create(ctx, &client.CreateQueueRequest{
		OwnerID:     generics.GetFakeObj[string](),
		OwnerType:   generics.GetFakeObj[string](),
		Namespace:   defaultNamespace,
		MaxInFlight: 5,
		MaxDepth:    100,
	})
	require.NoError(e.T(), err)
	require.NoError(e.T(), e.queueReady(ctx, queue.ID))

	key := "same-logical-event"
	first, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal: &example.ExampleSignal{
			Arg1: generics.GetFakeObj[string](),
			Arg2: generics.GetFakeObj[string](),
		},
		IdempotencyKey: key,
	})
	require.NoError(e.T(), err)
	require.False(e.T(), first.Deduplicated)
	e.waitForSignalStatus(ctx, first.ID, app.StatusSuccess)

	second, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal: &example.ExampleSignal{
			Arg1: generics.GetFakeObj[string](),
			Arg2: generics.GetFakeObj[string](),
		},
		IdempotencyKey: key,
	})
	require.NoError(e.T(), err)
	require.True(e.T(), second.Deduplicated)
	require.Equal(e.T(), first.ID, second.ID)
	require.Equal(e.T(), first.WorkflowID, second.WorkflowID)
	var queueSignal app.QueueSignal
	require.Eventually(e.T(), func() bool {
		queueSignal = app.QueueSignal{}
		err := e.service.DB.WithContext(ctx).Where(app.QueueSignal{ID: first.ID}).First(&queueSignal).Error
		return err == nil && queueSignal.Status.Status == app.StatusSuccess && queueSignal.ExecutionCount == 1
	}, 5*time.Second, 100*time.Millisecond)
	require.Equal(e.T(), 1, queueSignal.ExecutionCount)

	third, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal: &example.ExampleSignal{
			Arg1: generics.GetFakeObj[string](),
			Arg2: generics.GetFakeObj[string](),
		},
		IdempotencyKey: "different-logical-event",
	})
	require.NoError(e.T(), err)
	require.False(e.T(), third.Deduplicated)
	require.NotEqual(e.T(), first.ID, third.ID)
}

func (e *EnqueueTestSuite) TestPanickingSignalUpdatesDBStatus() {
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

	err = e.queueReady(ctx, q.ID)
	require.Nil(e.T(), err)

	// enqueue a panicking signal
	resp, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal: &example.PanickingSignal{
			Message: "test panic",
		},
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), resp)

	e.waitForSignalStatus(ctx, resp.ID, app.StatusError)

	// verify the DB has the error status persisted (not stuck in-progress)
	var qs app.QueueSignal
	res := e.service.DB.WithContext(ctx).First(&qs, "id = ?", resp.ID)
	require.Nil(e.T(), res.Error)
	require.Equal(e.T(), app.StatusError, qs.Status.Status)
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

	err = e.queueReady(ctx, q.ID)
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

	e.waitForSignalStatus(ctx, resp.ID, app.StatusError)

	// verify the DB has the error status persisted
	var qs app.QueueSignal
	res := e.service.DB.WithContext(ctx).First(&qs, "id = ?", resp.ID)
	require.Nil(e.T(), res.Error)
	require.Equal(e.T(), app.StatusError, qs.Status.Status)
}
