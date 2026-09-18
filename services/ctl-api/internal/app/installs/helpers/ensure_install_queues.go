package helpers

import (
	"context"
	"fmt"

	pkgworkflows "github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
)

const componentHealthEvaluateEmitterName = "component-health-evaluate"

// EnsureInstallQueues creates the install queues if they don't already exist.
// Safe to call multiple times — queueClient.Create is idempotent.
// Also updates existing queue capacities if they have changed.
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

	ownerType := plugins.TableName(s.db, app.Install{})
	specs, ok := queuenames.Specs(queuenames.OwnerInstalls)
	if !ok {
		return fmt.Errorf("install queue specs are not registered")
	}
	for _, q := range specs {
		if q.Name == InstallComponentHealthQueueName {
			continue
		}
		namespace := installsNamespace
		switch q.Name {
		case InstallActionWorkflowsQueueName, InstallDriftWorkflowsQueueName,
			InstallActionCronSignalsQueueName, InstallDriftCronSignalsQueueName:
			namespace = cronsNamespace
		}
		existing, err := s.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
			OwnerID:     installID,
			OwnerType:   ownerType,
			Namespace:   namespace,
			Name:        q.Name,
			MaxInFlight: q.MaxInFlight,
			MaxDepth:    q.MaxDepth,
		})
		if err != nil {
			return fmt.Errorf("unable to ensure %s queue: %w", q.Name, err)
		}

		// queueClient.Create migrates the queue workflow when its namespace
		// changed; migrate the queue's emitters to match (no-op if unchanged).
		if err := s.emitterClient.MigrateQueueEmitters(ctx, existing.ID, namespace); err != nil {
			return fmt.Errorf("unable to migrate %s queue emitters: %w", q.Name, err)
		}
	}

	if err := s.ensureComponentHealthQueue(ctx, installID, ownerType, cronsNamespace); err != nil {
		return err
	}

	return nil
}

// ensureComponentHealthQueue creates the component-health queue, which
// serializes evaluation for one install.
//
// It deliberately creates no cron emitter. Evaluation is driven by the runner's
// report and, for installs that go quiet, by one fleet-wide sweep — an emitter
// per install cost a workflow execution a minute forever and grew 1:1 with the
// fleet. Any emitter left over from that design is removed here.
func (s *Helpers) ensureComponentHealthQueue(ctx context.Context, installID, ownerType, namespace string) error {
	spec, ok := queuenames.SpecByName(queuenames.OwnerInstalls, queuenames.InstallComponentHealthQueueName)
	if !ok {
		return fmt.Errorf("install component health queue is not registered")
	}
	q, err := s.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     installID,
		OwnerType:   ownerType,
		Namespace:   namespace,
		Name:        spec.Name,
		MaxInFlight: spec.MaxInFlight,
		MaxDepth:    spec.MaxDepth,
	})
	if err != nil {
		return fmt.Errorf("unable to ensure %s queue: %w", InstallComponentHealthQueueName, err)
	}

	// Create migrates the queue workflow when its namespace changed; move the
	// emitter to match (no-op if unchanged).
	if err := s.emitterClient.MigrateQueueEmitters(ctx, q.ID, namespace); err != nil {
		return fmt.Errorf("unable to migrate %s queue emitters: %w", InstallComponentHealthQueueName, err)
	}

	emitters, err := s.emitterClient.GetEmittersByQueueID(ctx, q.ID)
	if err != nil {
		return fmt.Errorf("unable to list emitters for %s queue: %w", InstallComponentHealthQueueName, err)
	}
	for _, em := range emitters {
		if em.Name != componentHealthEvaluateEmitterName {
			continue
		}
		if err := s.emitterClient.DeleteEmitter(ctx, em.ID); err != nil {
			return fmt.Errorf("unable to remove legacy component health emitter: %w", err)
		}
	}

	return nil
}
