package helpers

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
)

func (h *Helpers) CreateAppSandboxQueue(ctx context.Context, appID string) error {
	_, err := h.ensureAppQueueByName(ctx, appID, queuenames.AppDefaultQueueName, false)
	return err
}
