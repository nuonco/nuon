package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	queuesignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// EnqueueSettingsChanged enqueues a settings_changed v2 signal to each
// process queue owned by the given runner.
func (h *Helpers) EnqueueSettingsChanged(ctx context.Context, runnerID string) error {
	var queues []app.Queue
	if err := h.db.WithContext(ctx).
		Where("owner_id = ? AND owner_type = ? AND name LIKE ?", runnerID, "runners", "runner-process-%").
		Find(&queues).Error; err != nil {
		return fmt.Errorf("unable to get queues for runner %s: %w", runnerID, err)
	}

	for _, q := range queues {
		if _, err := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
			QueueID: q.ID,
			Signal: queuesignal.NewRaw("settings_changed", map[string]any{
				"runner_id": runnerID,
			}),
		}); err != nil {
			return fmt.Errorf("unable to enqueue settings_changed to queue %s: %w", q.ID, err)
		}
	}

	return nil
}
