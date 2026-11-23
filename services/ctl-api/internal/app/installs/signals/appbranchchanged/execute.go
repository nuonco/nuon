package appbranchchanged

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/updateappconfig"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
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

	if latest.DesiredOnly {
		if _, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   s.InstallID,
			OwnerType: "installs",
			QueueName: queuenames.InstallSignalsQueueName,
			Signal: &updateappconfig.Signal{
				InstallID:      s.InstallID,
				NewAppConfigID: latest.AppConfigID,
				TriggeredBy:    "app-branch",
				InstallGroupID: latest.InstallGroupID,
				Metadata:       map[string]string{"source": "app-branch", "rollout": "desired-only"},
			},
		}); err != nil {
			return fmt.Errorf("unable to enqueue update-app-config signal: %w", err)
		}
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
