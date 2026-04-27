package helpers

import (
	"context"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// enqueueAppConfigUpdated enqueues the app_config_updated signal on the install's
// install-signals queue. This triggers emitter reconciliation for action cron
// triggers and drift detection.
func (s *Helpers) enqueueAppConfigUpdated(ctx context.Context, installID string) {
	l := zap.L()

	var queue app.Queue
	if err := s.db.WithContext(ctx).
		Where(app.Queue{OwnerID: installID, Name: InstallSignalsQueueName}).
		First(&queue).Error; err != nil {
		l.Warn("unable to find install-signals queue for app_config_updated",
			zap.String("install_id", installID),
			zap.Error(err),
		)
		return
	}

	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal: signal.NewRaw("app_config_updated", map[string]any{
			"install_id": installID,
		}),
	}); err != nil {
		l.Warn("unable to enqueue app_config_updated signal",
			zap.String("install_id", installID),
			zap.Error(err),
		)
	}
}
