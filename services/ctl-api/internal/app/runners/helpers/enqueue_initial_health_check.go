package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	queuesignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// MaybeEnqueueInitialHealthCheck ensures the process has an initial health check queued.
func (h *Helpers) MaybeEnqueueInitialHealthCheck(ctx context.Context, runnerID string, process *app.RunnerProcess) error {
	if process.InitialHealthCheck {
		return nil
	}

	sweeps, err := h.featuresClient.OrgHealthcheckSweepsEnabled(ctx, process.OrgID)
	if err != nil {
		return fmt.Errorf("unable to evaluate org healthcheck sweeps flag: %w", err)
	}
	if sweeps {
		return nil
	}

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

	expiresAt := time.Now().Add(1 * time.Hour)
	if _, err := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   q.ID,
		OwnerID:   process.ID,
		OwnerType: plugins.TableName(h.db, app.RunnerProcess{}),
		ExpiresAt: &expiresAt,
		Signal: queuesignal.NewRaw("process_healthcheck", map[string]any{
			"runner_id":  runnerID,
			"process_id": process.ID,
		}),
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
