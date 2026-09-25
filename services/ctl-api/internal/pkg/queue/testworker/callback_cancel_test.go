package testworker

import (
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/example"
)

// TestParentNeverCarriesOnAfterChildCancelled runs the real queue worker and
// asserts the product invariant behind step/workflow cancellation: a parent
// awaiting a child's completion callback must never carry on when that child
// is cancelled. The parent signal must terminate as an error, not success.
func (e *EnqueueTestSuite) TestParentNeverCarriesOnAfterChildCancelled() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())

	q, err := e.service.Client.Create(ctx, &client.CreateQueueRequest{
		OwnerID:     generics.GetFakeObj[string](),
		OwnerType:   generics.GetFakeObj[string](),
		Namespace:   defaultNamespace,
		MaxInFlight: 5,
		MaxDepth:    100,
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), q)
	require.Nil(e.T(), e.service.Client.QueueReady(ctx, q.ID))

	// Enqueue the parent, which blocks in execute awaiting a completion callback.
	awaitID := generics.GetFakeObj[string]()
	parentResp, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal:  &example.AwaitCallbackSignal{AwaitID: awaitID},
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), parentResp)

	pollTimeout := 15 * time.Second
	require.Eventually(e.T(), func() bool {
		var qs app.QueueSignal
		res := e.service.DB.WithContext(ctx).First(&qs, "id = ?", parentResp.ID)
		return res.Error == nil && qs.Status.Status == app.StatusInProgress
	}, pollTimeout, 200*time.Millisecond)

	childResp, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal:  &example.CancellableSignal{},
		Callback: callback.Ref{
			WorkflowID: parentResp.WorkflowID,
			SignalName: callback.SignalName(awaitID),
			Namespace:  defaultNamespace,
		},
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), childResp)

	require.Eventually(e.T(), func() bool {
		var qs app.QueueSignal
		res := e.service.DB.WithContext(ctx).First(&qs, "id = ?", childResp.ID)
		return res.Error == nil && qs.Status.Status == app.StatusInProgress
	}, pollTimeout, 200*time.Millisecond)

	// Cancel the child while the parent is awaiting its callback.
	cancelResp, err := e.service.Client.CancelSignal(ctx, childResp.ID)
	require.Nil(e.T(), err)
	require.NotNil(e.T(), cancelResp)

	require.Eventually(e.T(), func() bool {
		var qs app.QueueSignal
		res := e.service.DB.WithContext(ctx).First(&qs, "id = ?", childResp.ID)
		return res.Error == nil && qs.Status.Status == app.StatusCancelled
	}, pollTimeout, 200*time.Millisecond)

	var parentStatus app.Status
	require.Eventually(e.T(), func() bool {
		var qs app.QueueSignal
		res := e.service.DB.WithContext(ctx).First(&qs, "id = ?", parentResp.ID)
		if res.Error != nil {
			return false
		}
		parentStatus = qs.Status.Status
		return parentStatus == app.StatusError || parentStatus == app.StatusSuccess
	}, pollTimeout, 200*time.Millisecond)
	require.Equalf(e.T(), app.StatusError, parentStatus,
		"parent carried on after child cancellation (status=%s)", parentStatus)

	// The child's terminal state must be stable. A late transport-layer
	// error/success stamp overwriting cancelled is exactly the
	// transport-vs-domain conflation this test guards against, so re-check
	// after the dispatcher has fully settled.
	time.Sleep(2 * time.Second)
	var childQS app.QueueSignal
	res := e.service.DB.WithContext(ctx).First(&childQS, "id = ?", childResp.ID)
	require.Nil(e.T(), res.Error)
	require.Equalf(e.T(), app.StatusCancelled, childQS.Status.Status,
		"child cancelled status was overwritten after settling (status=%s)", childQS.Status.Status)
}

// TestParentCarriesOnAfterChildSuccess is the positive control for the
// cancellation test above: with identical callback wiring, a successful child
// lets the parent complete successfully. This proves the cancelled case fails
// because of the cancellation semantics, not broken wiring.
func (e *EnqueueTestSuite) TestParentCarriesOnAfterChildSuccess() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())

	q, err := e.service.Client.Create(ctx, &client.CreateQueueRequest{
		OwnerID:     generics.GetFakeObj[string](),
		OwnerType:   generics.GetFakeObj[string](),
		Namespace:   defaultNamespace,
		MaxInFlight: 5,
		MaxDepth:    100,
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), q)
	require.Nil(e.T(), e.service.Client.QueueReady(ctx, q.ID))

	awaitID := generics.GetFakeObj[string]()
	parentResp, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal:  &example.AwaitCallbackSignal{AwaitID: awaitID},
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), parentResp)

	pollTimeout := 15 * time.Second
	require.Eventually(e.T(), func() bool {
		var qs app.QueueSignal
		res := e.service.DB.WithContext(ctx).First(&qs, "id = ?", parentResp.ID)
		return res.Error == nil && qs.Status.Status == app.StatusInProgress
	}, pollTimeout, 200*time.Millisecond)

	childResp, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal: &example.ExampleSignal{
			Arg1: generics.GetFakeObj[string](),
			Arg2: generics.GetFakeObj[string](),
		},
		Callback: callback.Ref{
			WorkflowID: parentResp.WorkflowID,
			SignalName: callback.SignalName(awaitID),
			Namespace:  defaultNamespace,
		},
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), childResp)

	require.Eventually(e.T(), func() bool {
		var qs app.QueueSignal
		res := e.service.DB.WithContext(ctx).First(&qs, "id = ?", parentResp.ID)
		return res.Error == nil && qs.Status.Status == app.StatusSuccess
	}, pollTimeout, 200*time.Millisecond)
}
