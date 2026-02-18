package deploygrouptoqueue

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	// Get the install group
	group, err := activities.AwaitGetInstallGroupByID(ctx, s.InstallGroupID)
	if err != nil {
		return fmt.Errorf("unable to get install group: %w", err)
	}

	logger.Info("deploying install group to queue",
		"install_group_id", group.ID,
		"install_group_name", group.Name,
		"install_count", len(group.InstallIDs),
		"max_parallel", group.MaxParallel,
		"requires_approval", group.RequiresApproval,
		"rollback_on_failure", group.RollbackOnFailure,
	)

	// TODO: Implement the actual deployment logic
	// This will involve:
	// 1. For each install in group.InstallIDs:
	//    a. Create a nested queue for the install deployment
	//    b. Enqueue install deployment signal with max_parallel concurrency
	// 2. If requires_approval is true:
	//    a. Wait for approval signal before proceeding
	// 3. Monitor deployments and handle failures:
	//    a. If rollback_on_failure is true and any deployment fails:
	//       - Trigger rollback for all installs in this group
	//       - Return error to stop workflow
	// 4. Track overall progress and success/failure

	// For now, just log that we would deploy
	logger.Info("would deploy installs in group",
		"install_group_id", group.ID,
		"install_ids", group.InstallIDs,
	)

	logger.Info("install group deployment completed",
		"install_group_id", group.ID,
	)

	return nil
}
