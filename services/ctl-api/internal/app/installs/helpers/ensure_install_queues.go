package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// EnsureInstallQueues creates the three install queues (workflows, signals, workflow-steps)
// if they don't already exist. If they exist, restarts the queue workflows.
func (s *Helpers) EnsureInstallQueues(ctx context.Context, installID string) error {
	queues := []struct {
		Name        string
		MaxInFlight int
	}{
		{InstallWorkflowsQueueName, 10},
		{InstallSignalsQueueName, 20},
		{InstallWorkflowStepsQueueName, 10},
	}

	ownerType := plugins.TableName(s.db, app.Install{})

	for _, q := range queues {
		var existing app.Queue
		if res := s.db.WithContext(ctx).
			Where(app.Queue{OwnerID: installID, Name: q.Name}).
			First(&existing); res.Error == nil {
			if err := s.queueClient.Restart(ctx, existing.ID); err != nil {
				return fmt.Errorf("unable to restart %s queue: %w", q.Name, err)
			}
			continue
		}

		_, err := s.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
			OwnerID:     installID,
			OwnerType:   ownerType,
			Namespace:   "installs",
			Name:        q.Name,
			MaxInFlight: q.MaxInFlight,
			MaxDepth:    50,
		})
		if err != nil {
			return fmt.Errorf("unable to create %s queue: %w", q.Name, err)
		}
	}

	return nil
}
