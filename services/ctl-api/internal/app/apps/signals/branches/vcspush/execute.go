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
		AppBranchID:         s.AppBranchID,
		AppBranchConfigID:   s.AppBranchConfigID,
		PlanOnly:            s.PlanOnly,
		EventType:           s.EventType,
		PRNumber:            s.PRNumber,
		HeadSHA:             s.HeadSHA,
		HeadRef:             s.HeadRef,
		BaseBranch:          s.BaseBranch,
		BaseSHA:             s.BaseSHA,
		ChangedFiles:        s.ChangedFiles,
		PusherEmails:        s.PusherEmails,
		SenderLogin:         s.SenderLogin,
		FallbackCreatedByID: s.FallbackCreatedByID,
		Draft:               s.Draft,
	})
	if err != nil {
		return fmt.Errorf("unable to trigger app branch run from vcs push: %w", err)
	}

	if resp.RunID == "" {
		logger.Info("app branch run skipped from vcs push",
			"app_branch_id", s.AppBranchID,
		)
		return nil
	}

	logger.Info("app branch run triggered from vcs push",
		"run_id", resp.RunID,
		"workflow_id", resp.WorkflowID,
		"queue_signal_id", resp.QueueSignalID,
	)

	return nil
}
