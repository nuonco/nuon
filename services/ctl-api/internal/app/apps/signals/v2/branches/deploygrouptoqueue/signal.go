package deploygrouptoqueue

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "app-branch-deploy-group-to-queue"

type Signal struct {
	InstallGroupID string `json:"install_group_id" validate:"required"`
	AppBranchID    string `json:"app_branch_id" validate:"required"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	// Use playground validator for struct tag validation
	v := validator.New()
	if err := v.Struct(s); err != nil {
		return errors.Wrap(err, "validation failed")
	}

	// Validate app branch exists
	_, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return errors.Wrap(err, "app branch not found")
	}

	// Validate install group exists
	_, err = activities.AwaitGetInstallGroupByID(ctx, s.InstallGroupID)
	if err != nil {
		return errors.Wrap(err, "install group not found")
	}

	return nil
}

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
