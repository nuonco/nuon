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
// Also updates MaxInFlight on existing queues if it has changed.
func (s *Helpers) EnsureInstallQueues(ctx context.Context, installID string) error {
	queues := []struct {
		Name        string
		MaxInFlight int
	}{
		{InstallWorkflowsQueueName, 25},
		{InstallSignalsQueueName, 20},
		{InstallWorkflowStepGroupsQueueName, 40},
		{InstallWorkflowStepsQueueName, 40},
		{InstallStateManagerQueueName, 5},
		{InstallGenerateStepsQueueName, 10},
		{InstallActionWorkflowsQueueName, 10},
		{InstallDriftWorkflowsQueueName, 5},
		{InstallActionCronSignalsQueueName, 10},
	}

	ownerType := plugins.TableName(s.db, app.Install{})

	for _, q := range queues {
		existing, err := s.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
			OwnerID:     installID,
			OwnerType:   ownerType,
			Namespace:   "installs",
			Name:        q.Name,
			MaxInFlight: q.MaxInFlight,
			MaxDepth:    50,
		})
		if err != nil {
			return fmt.Errorf("unable to ensure %s queue: %w", q.Name, err)
		}

		// Update MaxInFlight if it has drifted from the desired value.
		if existing.MaxInFlight != q.MaxInFlight {
			s.db.WithContext(ctx).Model(existing).Update("max_in_flight", q.MaxInFlight)
		}
	}

	return nil
}
