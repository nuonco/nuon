package appbranchchanged

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	if err := activities.AwaitEnsureInstallAppBranch(ctx, &activities.EnsureInstallAppBranchInput{
		InstallID:   s.InstallID,
		AppBranchID: s.AppBranchID,
	}); err != nil {
		return fmt.Errorf("unable to update app branch: %w", err)
	}

	latest, err := activities.AwaitGetLatestAppBranchRunForInstall(ctx, &activities.GetLatestAppBranchRunForInstallInput{
		AppBranchID: s.AppBranchID,
		InstallID:   s.InstallID,
	})
	if err != nil {
		return fmt.Errorf("unable to get latest app branch run: %w", err)
	}
	if latest.AppConfigID == "" || latest.AlreadyCurrent {
		return nil
	}

	installGroupID := s.InstallGroupID
	if installGroupID == "" {
		installGroupID = latest.InstallGroupID
	}
	cb := callback.New(ctx, s.InstallID)
	if _, err := activities.AwaitCreateInstallAppConfigVersionWorkflow(ctx, &activities.CreateInstallAppConfigVersionWorkflowInput{
		InstallID:      s.InstallID,
		NewAppConfigID: latest.AppConfigID,
		AppBranchRunID: latest.AppBranchRunID,
		InstallGroupID: installGroupID,
		Callback:       cb,
	}); err != nil {
		return fmt.Errorf("unable to create install app config workflow: %w", err)
	}

	if _, err := callback.AwaitWithTimeout(ctx, cb, callback.FallbackAwaitTimeout); err != nil {
		return fmt.Errorf("install app config workflow failed: %w", err)
	}
	return nil
}
