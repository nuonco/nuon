package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	queuesignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const runnerSignalsQueueName = "runner-signals"

// EnsureRunnerSignalsQueue creates the runner-signals queue if it doesn't exist.
// Safe to call multiple times — queueClient.Create is idempotent.
func (h *Helpers) EnsureRunnerSignalsQueue(ctx context.Context, runnerID string) error {
	q, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     runnerID,
		OwnerType:   "runners",
		Namespace:   "runners",
		Name:        runnerSignalsQueueName,
		MaxInFlight: 10,
		MaxDepth:    50,
	})
	if err != nil {
		return fmt.Errorf("unable to ensure runner-signals queue: %w", err)
	}

	emitters, err := h.emitterClient.GetEmittersByQueueID(ctx, q.ID)
	if err != nil {
		return fmt.Errorf("unable to list emitters for runner-signals queue: %w", err)
	}

	hasHealthcheckEmitter := false
	for _, em := range emitters {
		if em.Name == "runner-healthcheck" {
			hasHealthcheckEmitter = true
			break
		}
	}

	if !hasHealthcheckEmitter {
		if _, err := h.emitterClient.CreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
			QueueID:         q.ID,
			Name:            "runner-healthcheck",
			Description:     "Periodic runner-level health check",
			Mode:            app.QueueEmitterModeCron,
			CronSchedule:    runnerHealthcheckSchedule(h.cfg.Env),
			JitterWindow:    30 * time.Second,
			SignalType:      "runner_healthcheck",
			SignalExpiresIn: runnerHealthcheckSignalExpiry,
			SignalTemplate: queuesignal.NewRaw("runner_healthcheck", map[string]any{
				"runner_id": runnerID,
			}),
		}); err != nil {
			return fmt.Errorf("unable to create runner healthcheck emitter: %w", err)
		}
	}

	return nil
}

// EnsureRunnerJobGroupQueues creates one queue per job group for the runner.
// Health monitoring is handled by process_healthcheck emitters on per-process queues.
// Safe to call multiple times — queueClient.Create is idempotent.
func (h *Helpers) EnsureRunnerJobGroupQueues(ctx context.Context, runner *app.Runner, settings *app.RunnerGroupSettings) error {
	for _, group := range allRunnerJobGroups {
		if _, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
			OwnerID:     runner.ID,
			OwnerType:   "runners",
			Namespace:   "runners",
			Name:        string(group),
			MaxInFlight: settings.MaxInFlightForGroup(group),
			MaxDepth:    100,
		}); err != nil {
			return fmt.Errorf("unable to ensure queue for job group %s: %w", group, err)
		}
	}

	return nil
}
