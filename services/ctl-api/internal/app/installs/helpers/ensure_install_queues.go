package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	queuesignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const (
	componentHealthEvaluateEmitterName  = "component-health-evaluate"
	componentHealthEvaluateSignalExpiry = 2 * time.Minute
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

	if err := s.ensureComponentHealthQueue(ctx, installID, ownerType); err != nil {
		return err
	}

	return nil
}

// ensureComponentHealthQueue creates the component-health queue and its cron
// emitter. The evaluator itself no-ops for orgs without the component-health
// feature, so the emitter is created unconditionally and enabling the feature
// needs no backfill.
func (s *Helpers) ensureComponentHealthQueue(ctx context.Context, installID, ownerType string) error {
	q, err := s.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     installID,
		OwnerType:   ownerType,
		Namespace:   "installs",
		Name:        InstallComponentHealthQueueName,
		MaxInFlight: 1,
		MaxDepth:    10,
	})
	if err != nil {
		return fmt.Errorf("unable to ensure %s queue: %w", InstallComponentHealthQueueName, err)
	}

	emitters, err := s.emitterClient.GetEmittersByQueueID(ctx, q.ID)
	if err != nil {
		return fmt.Errorf("unable to list emitters for %s queue: %w", InstallComponentHealthQueueName, err)
	}
	for _, em := range emitters {
		if em.Name == componentHealthEvaluateEmitterName {
			return nil
		}
	}

	if _, err := s.emitterClient.CreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
		QueueID:         q.ID,
		Name:            componentHealthEvaluateEmitterName,
		Description:     "Periodic component health evaluation",
		Mode:            app.QueueEmitterModeCron,
		CronSchedule:    "* * * * *",
		JitterWindow:    30 * time.Second,
		SignalType:      "component-health-evaluate",
		SignalExpiresIn: componentHealthEvaluateSignalExpiry,
		SignalTemplate: queuesignal.NewRaw("component-health-evaluate", map[string]any{
			"install_id": installID,
		}),
	}); err != nil {
		return fmt.Errorf("unable to create component health evaluate emitter: %w", err)
	}

	return nil
}
