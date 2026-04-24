package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	queuesignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// EnqueueProcessShutdown enqueues the process_shutdown signal to the process's
// queue and stops the queue emitters (health check cron, uptime check) so they
// no longer fire for this process.
func (h *Helpers) EnqueueProcessShutdown(ctx context.Context, process *app.RunnerProcess) error {
	q, err := h.queueClient.GetQueueByOwnerAndName(ctx, process.RunnerID, "runners", fmt.Sprintf("runner-process-%s", process.ID))
	if err != nil {
		return fmt.Errorf("unable to find process queue: %w", err)
	}

	// Stop emitters first so health checks stop firing
	if err := h.StopProcessQueue(ctx, q.ID); err != nil {
		return fmt.Errorf("unable to stop process queue emitters: %w", err)
	}

	// Enqueue the process_shutdown signal
	if _, err := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   q.ID,
		OwnerType: plugins.TableName(h.db, app.RunnerProcess{}),
		OwnerID:   process.ID,
		Signal: queuesignal.NewRaw("process_shutdown", map[string]any{
			"runner_id":  process.RunnerID,
			"process_id": process.ID,
		}),
	}); err != nil {
		return fmt.Errorf("unable to enqueue process_shutdown signal: %w", err)
	}

	return nil
}
