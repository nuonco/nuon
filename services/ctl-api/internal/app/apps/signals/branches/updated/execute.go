package updated

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/appbranchchanged"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	result, err := activities.AwaitReconcileAppBranchConfigInstalls(ctx, &activities.ReconcileAppBranchConfigInstallsInput{
		AppBranchID:       s.AppBranchID,
		AppBranchConfigID: s.AppBranchConfigID,
	})
	if err != nil {
		return fmt.Errorf("unable to reconcile app branch installs: %w", err)
	}

	for _, install := range result.Installs {
		if !install.ShouldSignal {
			continue
		}

		if _, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   install.InstallID,
			OwnerType: "installs",
			QueueName: queuenames.InstallSignalsQueueName,
			Signal: &appbranchchanged.Signal{
				InstallID:      install.InstallID,
				AppBranchID:    s.AppBranchID,
				InstallGroupID: install.InstallGroupID,
			},
		}); err != nil {
			return fmt.Errorf("unable to enqueue app branch change for install %s: %w", install.InstallID, err)
		}
	}

	return nil
}
