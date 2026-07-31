package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	queuesignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// EnqueueInstallSignal enqueues a v2 signal onto the named queue for an install.
func (h *Helpers) EnqueueInstallSignal(ctx context.Context, installID, queueName string, sig queuesignal.Signal) error {
	var q app.Queue
	if res := h.db.WithContext(ctx).
		Where(app.Queue{OwnerID: installID, Name: queueName}).
		First(&q); res.Error != nil {
		return fmt.Errorf("unable to find %s queue for install %s: %w", queueName, installID, res.Error)
	}

	_, err := h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal:  sig,
	})
	return err
}
