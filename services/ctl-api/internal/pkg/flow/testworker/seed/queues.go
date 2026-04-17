package seed

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type EnsureQueuesResult struct {
	PrimaryQueue *app.Queue // for the execute-flow signal itself
	StepQueue    *app.Queue // "install-workflow-steps"
	SignalQueue  *app.Queue // "install-signals"
}

const defaultNamespace = "default"

func (s *Seeder) EnsureQueues(ctx context.Context, t *testing.T, ownerID, ownerType string) *EnsureQueuesResult {
	result := &EnsureQueuesResult{}

	// Primary queue (unnamed) — where the execute-flow signal runs
	primary, err := s.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     ownerID,
		OwnerType:   ownerType,
		Namespace:   defaultNamespace,
		MaxInFlight: 5,
		MaxDepth:    100,
	})
	require.Nil(t, err)
	result.PrimaryQueue = primary

	// Step queue — where execute-workflow-step signals run
	stepQ, err := s.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     ownerID,
		OwnerType:   ownerType,
		Namespace:   defaultNamespace,
		Name:        "install-workflow-steps",
		MaxInFlight: 5,
		MaxDepth:    100,
	})
	require.Nil(t, err)
	result.StepQueue = stepQ

	// Signal queue — where inner step signals + generate-steps signals run
	sigQ, err := s.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     ownerID,
		OwnerType:   ownerType,
		Namespace:   defaultNamespace,
		Name:        "install-signals",
		MaxInFlight: 5,
		MaxDepth:    100,
	})
	require.Nil(t, err)
	result.SignalQueue = sigQ

	// Wait for all queues to be ready
	require.Nil(t, s.queueClient.QueueReady(ctx, primary.ID))
	require.Nil(t, s.queueClient.QueueReady(ctx, stepQ.ID))
	require.Nil(t, s.queueClient.QueueReady(ctx, sigQ.ID))

	return result
}
