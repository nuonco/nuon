package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
)

const runnerSignalsQueueName = "runner-signals"

// EnsureRunnerSignalsQueue creates the runner-signals queue if it doesn't exist.
// If it exists, restarts the queue workflow.
func (h *Helpers) EnsureRunnerSignalsQueue(ctx context.Context, runnerID string) error {
	var existing app.Queue
	if res := h.db.WithContext(ctx).
		Where(app.Queue{OwnerID: runnerID, Name: runnerSignalsQueueName}).
		First(&existing); res.Error == nil {
		return h.queueClient.Restart(ctx, existing.ID)
	}

	_, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     runnerID,
		OwnerType:   "runners",
		Namespace:   signals.TemporalNamespace,
		Name:        runnerSignalsQueueName,
		MaxInFlight: 10,
		MaxDepth:    50,
	})
	if err != nil {
		return fmt.Errorf("unable to create runner-signals queue: %w", err)
	}

	return nil
}

// EnsureRunnerJobGroupQueues creates one queue per job group for the runner if they don't exist.
// Also creates the healthcheck cron emitter on the health-check queue.
func (h *Helpers) EnsureRunnerJobGroupQueues(ctx context.Context, runner *app.Runner, settings *app.RunnerGroupSettings) error {
	var healthCheckQueueID string

	for _, group := range allRunnerJobGroups {
		var existing app.Queue
		if res := h.db.WithContext(ctx).
			Where(app.Queue{OwnerID: runner.ID, Name: string(group)}).
			First(&existing); res.Error == nil {
			if err := h.queueClient.Restart(ctx, existing.ID); err != nil {
				return fmt.Errorf("unable to restart queue for job group %s: %w", group, err)
			}
			if group == app.RunnerJobGroupHealthChecks {
				healthCheckQueueID = existing.ID
			}
			continue
		}

		q, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
			OwnerID:     runner.ID,
			OwnerType:   "runners",
			Namespace:   signals.TemporalNamespace,
			Name:        string(group),
			MaxInFlight: settings.MaxInFlightForGroup(group),
			MaxDepth:    100,
		})
		if err != nil {
			return fmt.Errorf("unable to create queue for job group %s: %w", group, err)
		}

		if group == app.RunnerJobGroupHealthChecks {
			healthCheckQueueID = q.ID
		}
	}

	if healthCheckQueueID == "" {
		return fmt.Errorf("health check queue was not created")
	}

	// Ensure the healthcheck cron emitter exists
	emitterName := fmt.Sprintf("runner-%s-health-check", runner.ID)
	var existing app.QueueEmitter
	if res := h.db.WithContext(ctx).
		Where(app.QueueEmitter{QueueID: healthCheckQueueID, Name: emitterName}).
		First(&existing); res.Error == nil {
		return nil
	}

	if _, err := h.emitterClient.CreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
		QueueID:        healthCheckQueueID,
		Name:           emitterName,
		Description:    "Periodic runner health check",
		Mode:           app.QueueEmitterModeCron,
		CronSchedule:   "* * * * *",
		SignalType:     healthcheckSignalType,
		SignalTemplate: &healthcheckSignalTemplate{RunnerID: runner.ID},
	}); err != nil {
		return fmt.Errorf("unable to create health check emitter: %w", err)
	}

	return nil
}
