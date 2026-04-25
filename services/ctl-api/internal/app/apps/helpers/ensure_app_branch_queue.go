package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// EnsureAppBranchQueue creates a Temporal queue workflow for the given app branch.
// Safe to call multiple times — queueClient.Create is idempotent.
func (h *Helpers) EnsureAppBranchQueue(ctx context.Context, branchID string) error {
	_, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     branchID,
		OwnerType:   plugins.TableName(h.db, app.AppBranch{}),
		Namespace:   "apps",
		MaxInFlight: 2,
		MaxDepth:    50,
	})
	if err != nil {
		return fmt.Errorf("unable to ensure queue for app branch %s: %w", branchID, err)
	}
	return nil
}
