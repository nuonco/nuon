package helpers

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
)

// CreateAppSandboxQueue creates a Temporal queue for the app's sandbox build workflow.
// This enables sandbox-build signals to be enqueued and processed against the app.
func (h *Helpers) CreateAppSandboxQueue(ctx context.Context, appID string) error {
	_, err := h.ensureAppQueueByName(ctx, appID, queuenames.AppDefaultQueueName, false)
	return err
}
