package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// CreateAppSandboxQueue creates a Temporal queue for the app's sandbox build workflow.
// This enables sandbox-build signals to be enqueued and processed against the app.
func (h *Helpers) CreateAppSandboxQueue(ctx context.Context, appID string) error {
	_, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     appID,
		OwnerType:   plugins.TableName(h.db, app.App{}),
		Namespace:   "apps",
		MaxInFlight: 1,
		MaxDepth:    50,
	})
	if err != nil {
		return fmt.Errorf("unable to create sandbox queue for app %s: %w", appID, err)
	}
	return nil
}
