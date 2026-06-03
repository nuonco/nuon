package vcspush

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	logger.Info("triggering app branch run from vcs push",
		"app_branch_id", s.AppBranchID,
		"app_branch_config_id", s.AppBranchConfigID,
	)

	resp, err := activities.AwaitTriggerAppBranchRunFromVCSPush(ctx, activities.TriggerAppBranchRunFromVCSPushRequest{
		AppBranchID:       s.AppBranchID,
		AppBranchConfigID: s.AppBranchConfigID,
		PlanOnly:          s.PlanOnly,
		EventType:         s.EventType,
		PRNumber:          s.PRNumber,
		HeadSHA:           s.HeadSHA,
		BaseBranch:        s.BaseBranch,
	})
	if err != nil {
		return fmt.Errorf("unable to trigger app branch run from vcs push: %w", err)
	}

	logger.Info("app branch run triggered from vcs push",
		"run_id", resp.RunID,
		"workflow_id", resp.WorkflowID,
		"queue_signal_id", resp.QueueSignalID,
	)

	return nil
}
