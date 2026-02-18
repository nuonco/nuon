package checkchanges

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "app-branch-check-changes"

type Signal struct {
	AppBranchID string `json:"app_branch_id" validate:"required"`
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

	_, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return errors.Wrap(err, "app branch not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	panic("here testing")
	// Get app branch with latest config
	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	// Check if branch has a config with VCS settings
	if len(branch.Configs) == 0 || branch.Configs[0].ConnectedGithubVCSConfig == nil {
		workflow.GetLogger(ctx).Info("no VCS config found for app branch",
			"app_branch_id", branch.ID)
		return nil
	}

	// Get latest commit from VCS using the config's VCS config ID
	latestCommit, err := activities.AwaitGetLatestCommitFromVCSByVcsConfigID(ctx, branch.Configs[0].ConnectedGithubVCSConfig.ID)
	if err != nil {
		return fmt.Errorf("unable to get latest commit: %w", err)
	}

	// TODO: LastSyncedCommit field is commented out in AppBranch struct - needs to be re-enabled
	// Compare with last synced commit
	// if latestCommit != branch.LastSyncedCommit {
	// 	workflow.GetLogger(ctx).Info("changes detected",
	// 		"app_branch_id", branch.ID,
	// 		"latest_commit", latestCommit,
	// 		"last_synced_commit", branch.LastSyncedCommit)
	//
	// 	// TODO: Enqueue update-app-config signal
	// 	// This will be implemented when update-app-config signal is ready and queue enqueue is available
	// } else {
	// 	workflow.GetLogger(ctx).Info("no changes detected",
	// 		"app_branch_id", branch.ID,
	// 		"commit", latestCommit)
	// }

	workflow.GetLogger(ctx).Info("latest commit fetched",
		"app_branch_id", branch.ID,
		"latest_commit", latestCommit)

	return nil
}
