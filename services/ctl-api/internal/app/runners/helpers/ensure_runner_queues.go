package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

const runnerSignalsQueueName = "runner-signals"

// EnsureRunnerSignalsQueue creates the runner-signals queue if it doesn't exist.
// Safe to call multiple times — queueClient.Create is idempotent.
func (h *Helpers) EnsureRunnerSignalsQueue(ctx context.Context, runnerID string) error {
	_, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     runnerID,
		OwnerType:   "runners",
		Namespace:   signals.TemporalNamespace,
		Name:        runnerSignalsQueueName,
		MaxInFlight: 10,
		MaxDepth:    50,
	})
	if err != nil {
		return fmt.Errorf("unable to ensure runner-signals queue: %w", err)
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
			Namespace:   signals.TemporalNamespace,
			Name:        string(group),
			MaxInFlight: settings.MaxInFlightForGroup(group),
			MaxDepth:    100,
		}); err != nil {
			return fmt.Errorf("unable to ensure queue for job group %s: %w", group, err)
		}
	}

	return nil
}
