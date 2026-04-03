package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// EnqueueInitialHealthCheck enqueues an immediate health check signal to the process's queue
// and marks the process as having received its initial health check.
func (h *Helpers) EnqueueInitialHealthCheck(ctx context.Context, runnerID string, process *app.RunnerProcess) error {
	queueName := fmt.Sprintf("runner-process-%s", process.ID)

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
		Signal: &processHealthcheckSignalTemplate{
			RunnerID:  runnerID,
			ProcessID: process.ID,
		},
	}); err != nil {
		return fmt.Errorf("unable to enqueue initial health check signal: %w", err)
	}

	if res := h.db.WithContext(ctx).
		Model(&app.RunnerProcess{ID: process.ID}).
		Update("initial_health_check", true); res.Error != nil {
		return fmt.Errorf("unable to mark initial health check: %w", res.Error)
	}

	return nil
}
