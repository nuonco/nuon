package testworker

import (
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/example"
)

func (e *EnqueueTestSuite) TestRestartQueue() {
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

	err = e.service.Client.HintRestartSingle(ctx, queue.ID)
	require.Nil(e.T(), err)

	err = e.service.Client.QueueReady(ctx, queue.ID)
	require.Nil(e.T(), err)

	status, err = e.service.Client.GetQueueStatus(ctx, queue.ID)
	require.Nil(e.T(), err)
	require.True(e.T(), status.Ready)
	require.False(e.T(), status.Stopped)
}

func (e *EnqueueTestSuite) TestRestartDefersWhileSignalIsExecuting() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())

	originalHintPeriod := e.service.Config.QueueContinueAsNewHintPeriod
	originalDrainTimeout := e.service.Config.QueueDrainTimeout
	e.service.Config.QueueContinueAsNewHintPeriod = 100 * time.Millisecond
	e.service.Config.QueueDrainTimeout = 100 * time.Millisecond
	e.T().Cleanup(func() {
		e.service.Config.QueueContinueAsNewHintPeriod = originalHintPeriod
		e.service.Config.QueueDrainTimeout = originalDrainTimeout
	})

	q, err := e.service.Client.Create(ctx, &client.CreateQueueRequest{
		OwnerID:     generics.GetFakeObj[string](),
		OwnerType:   generics.GetFakeObj[string](),
		Namespace:   defaultNamespace,
		MaxInFlight: 5,
		MaxDepth:    100,
	})
	require.NoError(e.T(), err)
	require.NotNil(e.T(), q)
	require.Eventually(e.T(), func() bool {
		return e.service.Client.QueueReady(ctx, q.ID) == nil
	}, 30*time.Second, 100*time.Millisecond)

	resp, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal:  &example.SlowSignal{},
	})
	require.NoError(e.T(), err)
	require.NotNil(e.T(), resp)

	var originalDequeuedAt any
	require.Eventually(e.T(), func() bool {
		var queueSignal app.QueueSignal
		res := e.service.DB.WithContext(ctx).Where(app.QueueSignal{ID: resp.ID}).First(&queueSignal)
		if res.Error != nil || queueSignal.Status.Status != app.StatusInProgress {
			return false
		}
		dequeuedAt, dequeued := queueSignal.Status.Metadata["dequeued_at"]
		_, executing := queueSignal.Status.Metadata["execute_started_at"]
		if !dequeued || !executing {
			return false
		}
		originalDequeuedAt = dequeuedAt
		return true
	}, 10*time.Second, 100*time.Millisecond)

	require.NoError(e.T(), e.service.Client.HintRestartSingle(ctx, q.ID))
	canResp, err := e.service.Client.CheckCAN(ctx, q.ID)
	require.NoError(e.T(), err)
	require.False(e.T(), canResp.Restarting)
	require.Never(e.T(), func() bool {
		var queueSignal app.QueueSignal
		res := e.service.DB.WithContext(ctx).Where(app.QueueSignal{ID: resp.ID}).First(&queueSignal)
		if res.Error != nil || queueSignal.Status.Status != app.StatusInProgress {
			return true
		}
		return queueSignal.Status.Metadata["dequeued_at"] != originalDequeuedAt
	}, 3*time.Second, 100*time.Millisecond)

	_, err = e.service.Client.CancelSignal(ctx, resp.ID)
	require.NoError(e.T(), err)
	require.Eventually(e.T(), func() bool {
		return e.service.Client.QueueReady(ctx, q.ID) == nil
	}, 10*time.Second, 100*time.Millisecond)
}
