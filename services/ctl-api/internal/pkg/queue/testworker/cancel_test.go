package testworker

import (
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/example"
)

func (e *EnqueueTestSuite) TestCancelSignalDuringExecute() {
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

	err = e.queueReady(ctx, q.ID)
	require.Nil(e.T(), err)

	resp, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal:  &example.SlowSignal{},
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), resp)

	pollTimeout := integrationEventuallyTimeout
	require.Eventually(e.T(), func() bool {
		var qs app.QueueSignal
		res := e.service.DB.WithContext(ctx).First(&qs, "id = ?", resp.ID)
		_, executing := qs.Status.Metadata["execute_started_at"]
		return res.Error == nil && executing
	}, pollTimeout, 200*time.Millisecond)

	cancelResp, err := e.service.Client.CancelSignal(ctx, resp.ID)
	require.Nil(e.T(), err)
	require.NotNil(e.T(), cancelResp)

	e.waitForSignalStatus(ctx, resp.ID, app.StatusCancelled)
}

func (e *EnqueueTestSuite) TestCancelCallbackInvoked() {
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

	err = e.queueReady(ctx, q.ID)
	require.Nil(e.T(), err)

	resp, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal:  &example.CancellableSignal{},
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), resp)

	pollTimeout := integrationEventuallyTimeout
	require.Eventually(e.T(), func() bool {
		var qs app.QueueSignal
		res := e.service.DB.WithContext(ctx).First(&qs, "id = ?", resp.ID)
		_, executing := qs.Status.Metadata["execute_started_at"]
		return res.Error == nil && executing
	}, pollTimeout, 200*time.Millisecond)

	cancelResp, err := e.service.Client.CancelSignal(ctx, resp.ID)
	require.Nil(e.T(), err)
	require.NotNil(e.T(), cancelResp)

	require.Eventually(e.T(), func() bool {
		var qs app.QueueSignal
		res := e.service.DB.WithContext(ctx).First(&qs, "id = ?", resp.ID)
		return res.Error == nil &&
			qs.Status.Status == app.StatusCancelled &&
			qs.Status.StatusHumanDescription == example.CancelCallbackMarker
	}, pollTimeout, 200*time.Millisecond)
}

func statusHistoryContains(status app.CompositeStatus, expected app.Status, description string) bool {
	for _, entry := range status.History {
		if entry.Status == expected && (description == "" || entry.StatusHumanDescription == description) {
			return true
		}
	}
	return false
}

func (e *EnqueueTestSuite) TestCancelAlreadyFinishedSignal() {
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

	err = e.queueReady(ctx, q.ID)
	require.Nil(e.T(), err)

	resp, err := e.service.Client.EnqueueSignal(ctx, &client.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal: &example.ExampleSignal{
			Arg1: generics.GetFakeObj[string](),
			Arg2: generics.GetFakeObj[string](),
		},
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), resp)

	e.waitForSignalStatus(ctx, resp.ID, app.StatusSuccess)

	cancelResp, err := e.service.Client.CancelSignal(ctx, resp.ID)
	require.Nil(e.T(), err)
	require.NotNil(e.T(), cancelResp)

	var qs app.QueueSignal
	res := e.service.DB.WithContext(ctx).First(&qs, "id = ?", resp.ID)
	require.Nil(e.T(), res.Error)
	require.Equal(e.T(), app.StatusSuccess, qs.Status.Status)
}
