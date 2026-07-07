package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

const AppBranchSandboxBuildsQueueName = "app-branch-sandbox-builds"

// EnsureAppBranchQueue creates Temporal queue workflows for the given app branch.
// Safe to call multiple times — queueClient.Create is idempotent.
func (h *Helpers) EnsureAppBranchQueue(ctx context.Context, branchID string) error {
	ownerType := plugins.TableName(h.db, app.AppBranch{})

	_, err := h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     branchID,
		OwnerType:   ownerType,
		Namespace:   "apps",
		MaxInFlight: 2,
		MaxDepth:    50,
	})
	if err != nil {
		return fmt.Errorf("unable to ensure default queue for app branch %s: %w", branchID, err)
	}

	_, err = h.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     branchID,
		OwnerType:   ownerType,
		Namespace:   "apps",
		Name:        AppBranchSandboxBuildsQueueName,
		MaxInFlight: 2,
		MaxDepth:    50,
	})
	if err != nil {
		return fmt.Errorf("unable to ensure sandbox builds queue for app branch %s: %w", branchID, err)
	}

	return nil
}
