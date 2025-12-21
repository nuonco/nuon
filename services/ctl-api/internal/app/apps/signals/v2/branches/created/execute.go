package created

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	// Get the app branch
	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	// Initialize state
	workflow.GetLogger(ctx).Info("app branch created", "app_branch_id", branch.ID, "name", branch.Name)

	// TODO: Enqueue check-changes signal to start sync loop
	// This will be implemented when check-changes signal is ready

	return nil
}
