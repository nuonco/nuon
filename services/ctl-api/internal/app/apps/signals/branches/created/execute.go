package created

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	logger := workflow.GetLogger(ctx)
	logger.Info("app branch created",
		"app_branch_id", branch.ID,
		"name", branch.Name,
		"app_branch_config_id", s.AppBranchConfigID,
	)

	resp, err := activities.AwaitTriggerAppBranchRunFromCreated(ctx, activities.TriggerAppBranchRunFromCreatedRequest{
		AppBranchID:       s.AppBranchID,
		AppBranchConfigID: s.AppBranchConfigID,
	})
	if err != nil {
		return fmt.Errorf("unable to trigger first app branch run: %w", err)
	}

	if resp.Skipped {
		logger.Info("skipped first app branch run",
			"app_branch_id", s.AppBranchID,
			"reason", resp.Reason,
		)
		return nil
	}

	logger.Info("triggered first app branch run",
		"run_id", resp.RunID,
		"workflow_id", resp.WorkflowID,
	)
	return nil
}
