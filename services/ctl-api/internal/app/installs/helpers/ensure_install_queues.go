package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// EnsureInstallQueues creates the install queues if they don't already exist.
// Safe to call multiple times — queueClient.Create is idempotent.
func (s *Helpers) EnsureInstallQueues(ctx context.Context, installID string) error {
	queues := []struct {
		Name        string
		MaxInFlight int
	}{
		{InstallWorkflowsQueueName, 10},
		{InstallSignalsQueueName, 20},
		{InstallWorkflowStepGroupsQueueName, 10},
		{InstallWorkflowStepsQueueName, 10},
		{InstallStateManagerQueueName, 5},
		{InstallGenerateStepsQueueName, 10},
	}

	ownerType := plugins.TableName(s.db, app.Install{})

	for _, q := range queues {
		if _, err := s.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
			OwnerID:     installID,
			OwnerType:   ownerType,
			Namespace:   "installs",
			Name:        q.Name,
			MaxInFlight: q.MaxInFlight,
			MaxDepth:    50,
		}); err != nil {
			return fmt.Errorf("unable to ensure %s queue: %w", q.Name, err)
		}
	}

	return nil
}
