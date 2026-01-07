package checkchanges

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "app-branch-check-changes"

type Signal struct {
	AppBranchID string `json:"app_branch_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.AppBranchID == "" {
		return errors.New("app_branch_id is required")
	}

	_, err := activities.AwaitGetAppBranchByID(ctx, s.AppBranchID)
	if err != nil {
		return errors.Wrap(err, "app branch not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// Get app branch
	branch, err := activities.AwaitGetAppBranchByID(ctx, s.AppBranchID)
	if err != nil {
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	// Get latest commit from VCS
	latestCommit, err := activities.AwaitGetLatestCommitFromVCS(ctx, branch.ConnectedGithubVCSConfigID)
	if err != nil {
		return fmt.Errorf("unable to get latest commit: %w", err)
	}

	// Compare with last synced commit
	if latestCommit != branch.LastSyncedCommit {
		workflow.GetLogger(ctx).Info("changes detected",
			"app_branch_id", branch.ID,
			"latest_commit", latestCommit,
			"last_synced_commit", branch.LastSyncedCommit)

		// TODO: Enqueue update-app-config signal
		// This will be implemented when update-app-config signal is ready and queue enqueue is available
	} else {
		workflow.GetLogger(ctx).Info("no changes detected",
			"app_branch_id", branch.ID,
			"commit", latestCommit)
	}

	return nil
}
