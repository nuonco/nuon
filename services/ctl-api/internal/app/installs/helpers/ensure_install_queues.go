package helpers

import (
	"context"
	"fmt"

	pkgworkflows "github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// EnsureInstallQueues creates the install queues if they don't already exist.
// Safe to call multiple times — queueClient.Create is idempotent.
// Also updates MaxInFlight on existing queues if it has changed.
func (s *Helpers) EnsureInstallQueues(ctx context.Context, installID string) error {
	var install app.Install
	if res := s.db.WithContext(ctx).Where(app.Install{ID: installID}).First(&install); res.Error != nil {
		return fmt.Errorf("unable to get install: %w", res.Error)
	}

	const installsNamespace = "installs"
	cronsNamespace := installsNamespace
	isolated, err := s.featuresClient.OrgCronNamespaceIsolationEnabled(ctx, install.OrgID)
	if err != nil {
		return fmt.Errorf("unable to evaluate cron namespace isolation: %w", err)
	}
	if isolated {
		cronsNamespace = pkgworkflows.InstallCronsNamespace
	}

	queues := []struct {
		Name        string
		Namespace   string
		MaxInFlight int
	}{
		{InstallWorkflowsQueueName, installsNamespace, 25},
		{InstallSignalsQueueName, installsNamespace, 20},
		{InstallWorkflowStepGroupsQueueName, installsNamespace, 40},
		{InstallWorkflowStepsQueueName, installsNamespace, 40},
		{InstallStateManagerQueueName, installsNamespace, 5},
		{InstallGenerateStepsQueueName, installsNamespace, 10},
		{InstallActionWorkflowsQueueName, cronsNamespace, 10},
		{InstallDriftWorkflowsQueueName, cronsNamespace, 5},
		{InstallActionCronSignalsQueueName, cronsNamespace, 10},
		{InstallDriftCronSignalsQueueName, cronsNamespace, 5},
	}

	ownerType := plugins.TableName(s.db, app.Install{})

	for _, q := range queues {
		existing, err := s.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
			OwnerID:     installID,
			OwnerType:   ownerType,
			Namespace:   q.Namespace,
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

		// queueClient.Create migrates the queue workflow when its namespace
		// changed; migrate the queue's emitters to match (no-op if unchanged).
		if err := s.emitterClient.MigrateQueueEmitters(ctx, existing.ID, q.Namespace); err != nil {
			return fmt.Errorf("unable to migrate %s queue emitters: %w", q.Name, err)
		}
	}

	return nil
}
