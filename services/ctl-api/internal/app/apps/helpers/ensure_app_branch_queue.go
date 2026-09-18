package helpers

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
)

const AppBranchSandboxBuildsQueueName = queuenames.AppBranchSandboxBuildsQueueName

const AppBranchDefaultMaxInFlight = 25

// EnsureAppBranchQueue creates Temporal queue workflows for the given app branch.
// Safe to call multiple times — queueClient.Create is idempotent.
func (h *Helpers) EnsureAppBranchQueue(ctx context.Context, branchID string) error {
	return h.EnsureAppBranchQueues(ctx, branchID)
}
