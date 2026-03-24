package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// EnsureAppBranchQueue ensures a Temporal queue workflow is running for the given app branch.
// If no queue DB record exists, it creates one and starts the workflow.
// If the DB record exists but the workflow may not be running, it calls Restart which uses
// UpdateWithStart to start the workflow if needed.
func (h *Helpers) EnsureAppBranchQueue(ctx context.Context, branchID string) error {
	var existing app.Queue
	if res := h.db.WithContext(ctx).First(&existing, "owner_id = ?", branchID); res.Error == nil {
		// DB record exists but Temporal workflow may not be running — Restart uses
		// UpdateWithStart which starts the workflow if it doesn't exist.
		if err := h.queueClient.Restart(ctx, existing.ID); err != nil {
			return fmt.Errorf("unable to restart queue for app branch %s: %w", branchID, err)
		}
		return nil
	}

	_, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     branchID,
		OwnerType:   plugins.TableName(h.db, app.AppBranch{}),
		Namespace:   "apps",
		MaxInFlight: 2,
		MaxDepth:    50,
	})
	if err != nil {
		return fmt.Errorf("unable to create queue for app branch %s: %w", branchID, err)
	}
	return nil
}
