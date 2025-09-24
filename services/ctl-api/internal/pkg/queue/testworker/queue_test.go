package testworker

import (
	"github.com/stretchr/testify/require"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/queue/client"
)

func (e *EnqueueTestSuite) TestCreateQueue() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())

	// now, create the queue
	queue, err := e.service.Client.Create(ctx, &client.CreateQueueRequest{
		OwnerID:     generics.GetFakeObj[string](),
		OwnerType:   generics.GetFakeObj[string](),
		Namespace:   defaultNamespace,
		MaxInFlight: 5,
		MaxDepth:    100,
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), queue)
}

func (e *EnqueueTestSuite) TestAwaitReadyOK() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())

	// now, create the queue
	queue, err := e.service.Client.Create(ctx, &client.CreateQueueRequest{
		OwnerID:     generics.GetFakeObj[string](),
		OwnerType:   generics.GetFakeObj[string](),
		Namespace:   defaultNamespace,
		MaxInFlight: 5,
		MaxDepth:    100,
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), queue)

	// wait for the event loop to be ready
	err = e.service.Client.QueueReady(ctx, queue.ID)
	require.Nil(e.T(), err)
}

func (e *EnqueueTestSuite) TestAwaitStatusOK() {
	ctx := e.service.Seed.EnsureAccount(e.T().Context(), e.T())
	ctx = e.service.Seed.EnsureOrg(ctx, e.T())

	// now, create the queue
	queue, err := e.service.Client.Create(ctx, &client.CreateQueueRequest{
		OwnerID:     generics.GetFakeObj[string](),
		OwnerType:   generics.GetFakeObj[string](),
		Namespace:   defaultNamespace,
		MaxInFlight: 5,
		MaxDepth:    100,
	})
	require.Nil(e.T(), err)
	require.NotNil(e.T(), queue)

	// now, restart the queue
	err = e.service.Client.QueueReady(ctx, queue.ID)
	require.Nil(e.T(), err)
}
