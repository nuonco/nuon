package helpers

import (
	"context"
	"fmt"

	pkgworkflows "github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	queuesignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const (
	runnerSignalsQueueName = queuenames.RunnerSignalsQueueName

	// RunnerHealthcheckCronsQueueName hosts the per-runner healthcheck cron
	// emitter for orgs without the org-healthcheck-sweeps feature. Sweep-enabled
	// orgs neither create it nor keep it (the migration terminates it).
	RunnerHealthcheckCronsQueueName = queuenames.RunnerHealthcheckCronsQueueName

	RunnerHealthcheckEmitterName = "runner-healthcheck"
)

// EnsureRunnerSignalsQueue creates the runner-signals queue if it doesn't
// exist. For orgs without org-healthcheck-sweeps it also ensures the legacy
// per-runner healthcheck cron queue + emitter. Safe to call multiple times.
func (h *Helpers) EnsureRunnerSignalsQueue(ctx context.Context, runnerID string) error {
	var runner app.Runner
	if res := h.db.WithContext(ctx).Where(app.Runner{ID: runnerID}).First(&runner); res.Error != nil {
		return fmt.Errorf("unable to get runner: %w", res.Error)
	}

	spec, ok := queuenames.SpecByName(queuenames.OwnerRunners, queuenames.RunnerSignalsQueueName)
	if !ok {
		return fmt.Errorf("runner signals queue is not registered")
	}
	_, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     runnerID,
		OwnerType:   "runners",
		Namespace:   "runners",
		Name:        spec.Name,
		MaxInFlight: spec.MaxInFlight,
		MaxDepth:    spec.MaxDepth,
	})
	if err != nil {
		return fmt.Errorf("unable to ensure runner-signals queue: %w", err)
	}

	sweeps, err := h.featuresClient.OrgHealthcheckSweepsEnabled(ctx, runner.OrgID)
	if err != nil {
		return fmt.Errorf("unable to evaluate org healthcheck sweeps flag: %w", err)
	}
	if sweeps {
		return nil
	}

	return h.EnsureRunnerHealthcheckEmitter(ctx, &runner)
}

// EnsureRunnerHealthcheckEmitter creates the legacy per-runner healthcheck cron
// queue + emitter (pre-org-healthcheck-sweeps behavior). Idempotent.
func (h *Helpers) EnsureRunnerHealthcheckEmitter(ctx context.Context, runner *app.Runner) error {
	healthcheckNamespace := "runners"
	isolated, err := h.featuresClient.OrgCronNamespaceIsolationEnabled(ctx, runner.OrgID)
	if err != nil {
		return fmt.Errorf("unable to evaluate cron namespace isolation: %w", err)
	}
	if isolated {
		healthcheckNamespace = pkgworkflows.RunnerHealthcheckCronsNamespace
	}

	spec, ok := queuenames.SpecByName(queuenames.OwnerRunners, queuenames.RunnerHealthcheckCronsQueueName)
	if !ok {
		return fmt.Errorf("runner healthcheck queue is not registered")
	}
	healthcheckQueue, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     runner.ID,
		OwnerType:   "runners",
		Namespace:   healthcheckNamespace,
		Name:        spec.Name,
		MaxInFlight: spec.MaxInFlight,
		MaxDepth:    spec.MaxDepth,
	})
	if err != nil {
		return fmt.Errorf("unable to ensure runner-healthcheck-crons queue: %w", err)
	}

	if err := h.emitterClient.MigrateQueueEmitters(ctx, healthcheckQueue.ID, healthcheckNamespace); err != nil {
		return fmt.Errorf("unable to migrate runner healthcheck emitters: %w", err)
	}

	if _, err := h.emitterClient.CreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
		QueueID:         healthcheckQueue.ID,
		Name:            RunnerHealthcheckEmitterName,
		Description:     "Periodic runner-level health check",
		Mode:            app.QueueEmitterModeCron,
		CronSchedule:    RunnerHealthcheckSchedule(h.cfg.Env),
		JitterWindow:    runnerHealthcheckJitterWindow,
		SignalType:      "runner_healthcheck",
		SignalExpiresIn: runnerHealthcheckSignalExpiry,
		SignalTemplate: queuesignal.NewRaw("runner_healthcheck", map[string]any{
			"runner_id": runner.ID,
		}),
	}); err != nil {
		return fmt.Errorf("unable to create runner healthcheck emitter: %w", err)
	}

	return nil
}

// EnsureRunnerJobGroupQueues creates one queue per job group for the runner.
// Safe to call multiple times — queueClient.Create is idempotent.
func (h *Helpers) EnsureRunnerJobGroupQueues(ctx context.Context, runner *app.Runner, settings *app.RunnerGroupSettings) error {
	for _, spec := range queuenames.RunnerJobGroupSpecs() {
		group := app.RunnerJobGroup(spec.Name)
		maxInFlight := settings.MaxInFlightForGroup(group)
		_, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
			OwnerID:     runner.ID,
			OwnerType:   "runners",
			Namespace:   "runners",
			Name:        spec.Name,
			MaxInFlight: maxInFlight,
			MaxDepth:    spec.MaxDepth,
		})
		if err != nil {
			return fmt.Errorf("unable to ensure queue for job group %s: %w", group, err)
		}
	}

	return nil
}
