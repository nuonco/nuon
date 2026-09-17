package helpers

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const appBranchCreatedSignalType signal.SignalType = "app-branch-created"

type appBranchCreatedSignal struct {
	AppBranchID       string `json:"app_branch_id"`
	AppBranchConfigID string `json:"app_branch_config_id"`
}

func (s *appBranchCreatedSignal) Type() signal.SignalType           { return appBranchCreatedSignalType }
func (s *appBranchCreatedSignal) Validate(_ workflow.Context) error { return nil }
func (s *appBranchCreatedSignal) Execute(_ workflow.Context) error  { return nil }

// EnqueueAppBranchCreatedIfFirst enqueues app-branch-created when configID is
// the branch's first AppBranchConfig. Later configs are ignored. DedupeKey
// makes a retry of the first config a no-op.
func (h *Helpers) EnqueueAppBranchCreatedIfFirst(ctx context.Context, appBranchID, appBranchConfigID string) error {
	if appBranchID == "" {
		return fmt.Errorf("app_branch_id is required")
	}
	if appBranchConfigID == "" {
		return fmt.Errorf("app_branch_config_id is required")
	}

	var count int64
	if err := h.db.WithContext(ctx).
		Model(&app.AppBranchConfig{}).
		Where(app.AppBranchConfig{AppBranchID: appBranchID}).
		Count(&count).Error; err != nil {
		return fmt.Errorf("unable to count app branch configs: %w", err)
	}
	if count != 1 {
		return nil
	}

	queue, err := h.queueClient.GetQueueByOwnerAndName(ctx, appBranchID, "app_branches", "app-branch-signals")
	if err != nil {
		return fmt.Errorf("unable to find app branch queue: %w", err)
	}

	dedupeKey := fmt.Sprintf("app-branch-created:%s", appBranchID)
	_, err = h.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   queue.ID,
		OwnerID:   appBranchID,
		OwnerType: "app_branches",
		DedupeKey: &dedupeKey,
		Signal: &appBranchCreatedSignal{
			AppBranchID:       appBranchID,
			AppBranchConfigID: appBranchConfigID,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to enqueue app-branch-created: %w", err)
	}
	return nil
}
