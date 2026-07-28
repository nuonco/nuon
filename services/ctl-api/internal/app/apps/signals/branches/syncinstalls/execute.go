package syncinstalls

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/installconfigsync"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	installs, err := activities.AwaitGetInstallsForAppByAppID(ctx, branch.AppID)
	if err != nil {
		return fmt.Errorf("unable to get installs for app: %w", err)
	}

	if len(installs) == 0 {
		logger.Info("no installs found for app, skipping sync")
		s.updateStepStatus(ctx, app.StatusSuccess, "no installs found", nil)
		return nil
	}

	enqueued := 0
	for _, install := range installs {
		_, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   install.ID,
			OwnerType: "installs",
			QueueName: installhelpers.InstallSignalsQueueName,
			Signal: &installconfigsync.Signal{
				InstallID:         install.ID,
				AppBranchID:       s.AppBranchID,
				AppBranchConfigID: s.AppBranchConfigID,
				AppBranchRunID:    s.AppBranchRunID,
				CommitSHA:         s.CommitSHA,
				TriggeredBy:       s.TriggeredBy,
			},
		})
		if err != nil {
			logger.Warn("failed to enqueue install-config-sync for install",
				"install_id", install.ID,
				"install_name", install.Name,
				"error", err,
			)
			continue
		}
		enqueued++
	}

	desc := fmt.Sprintf("enqueued sync for %d installs", enqueued)
	s.updateStepStatus(ctx, app.StatusSuccess, desc, map[string]any{
		"enqueued": enqueued,
		"total":    len(installs),
	})

	logger.Info(desc)
	return nil
}

func (s *Signal) updateStepStatus(ctx workflow.Context, status app.Status, desc string, metadata map[string]any) {
	if s.StepID == "" {
		return
	}

	_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.StepID,
		Status: app.CompositeStatus{
			Status:                 status,
			StatusHumanDescription: desc,
			Metadata:               metadata,
		},
	})
}
