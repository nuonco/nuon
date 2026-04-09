package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	queuesignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// MaybeEnqueueInitialHealthCheck fetches the process, checks if it has already
// received its initial health check, and if not enqueues one and updates the flag.
func (h *Helpers) MaybeEnqueueInitialHealthCheck(ctx context.Context, runnerID, processID string) error {
	var process app.RunnerProcess
	if res := h.db.WithContext(ctx).First(&process, "id = ?", processID); res.Error != nil {
		return fmt.Errorf("unable to fetch process %s: %w", processID, res.Error)
	}

	if process.InitialHealthCheck {
		return nil
	}

	queueName := fmt.Sprintf("runner-process-%s", processID)

	var q app.Queue
	if res := h.db.WithContext(ctx).
		Where(&app.Queue{
			OwnerID:   runnerID,
			OwnerType: "runners",
			Name:      queueName,
		}).First(&q); res.Error != nil {
		return fmt.Errorf("unable to find process queue %s: %w", queueName, res.Error)
	}

	if _, err := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal: queuesignal.NewRaw("process_healthcheck", map[string]any{
			"runner_id":  runnerID,
			"process_id": processID,
		}),
	}); err != nil {
		return fmt.Errorf("unable to enqueue initial health check signal: %w", err)
	}

	if res := h.db.WithContext(ctx).
		Model(&app.RunnerProcess{ID: processID}).
		Update("initial_health_check", true); res.Error != nil {
		return fmt.Errorf("unable to mark initial health check: %w", res.Error)
	}

	return nil
}
