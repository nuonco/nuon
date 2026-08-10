package syncinstalls

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	branchactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	syncID := s.AppInstallConfigSyncID
	if syncID == "" {
		result, err := branchactivities.AwaitCreateAppInstallConfigSync(ctx, &branchactivities.CreateAppInstallConfigSyncInput{
			AppID:       s.AppID,
			CommitSHA:   s.CommitSHA,
			TriggeredBy: s.TriggeredBy,
			CreatedByID: s.FallbackCreatedByID,
		})
		if err != nil {
			return fmt.Errorf("unable to create app install config sync: %w", err)
		}
		syncID = result.ID
	}

	_ = branchactivities.AwaitUpdateAppInstallConfigSyncStatus(ctx, &branchactivities.UpdateAppInstallConfigSyncStatusInput{
		ID:                syncID,
		Status:            string(app.StatusInProgress),
		StatusDescription: "syncing",
	})

	installsConfig, err := branchactivities.AwaitGetAppInstallsConfigByAppID(ctx, s.AppID)
	if err != nil {
		return fmt.Errorf("unable to get app installs config: %w", err)
	}

	if !installsConfig.Found {
		logger.Info("no installs config found for app, skipping sync")
		_ = branchactivities.AwaitUpdateAppInstallConfigSyncStatus(ctx, &branchactivities.UpdateAppInstallConfigSyncStatusInput{
			ID:                syncID,
			Status:            string(app.StatusSuccess),
			StatusDescription: "no installs config found",
		})
		s.updateStepStatus(ctx, app.StatusSuccess, "no installs config found", nil)
		return nil
	}

	installsDir := installsConfig.Directory
	if installsDir == "" {
		installsDir = "."
	}

	metadata := map[string]string{
		"app_id":             s.AppID,
		"sync_id":            syncID,
		"commit_sha":         s.CommitSHA,
		"triggered_by":       s.TriggeredBy,
		"installs_directory": installsDir,
	}

	wfResult, err := branchactivities.AwaitCreateInstallSyncWorkflow(ctx, &branchactivities.CreateInstallSyncWorkflowInput{
		AppID:                  s.AppID,
		AppInstallConfigSyncID: syncID,
		Metadata:               metadata,
	})
	if err != nil {
		_ = branchactivities.AwaitUpdateAppInstallConfigSyncStatus(ctx, &branchactivities.UpdateAppInstallConfigSyncStatusInput{
			ID:                syncID,
			Status:            string(app.StatusError),
			StatusDescription: "failed to create workflow",
		})
		return fmt.Errorf("unable to create install sync workflow: %w", err)
	}

	logger.Info("created install sync workflow",
		"workflow_id", wfResult.WorkflowID,
		"sync_id", syncID,
	)

	cb := callback.New(ctx, syncID)
	_, err = sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         s.AppID,
		OwnerType:       "apps",
		SignalOwnerID:   wfResult.WorkflowID,
		SignalOwnerType: "install_workflows",
		Signal: &executeflow.Signal{
			WorkflowID: wfResult.WorkflowID,
		},
		Callback: cb,
	})
	if err != nil {
		_ = branchactivities.AwaitUpdateAppInstallConfigSyncStatus(ctx, &branchactivities.UpdateAppInstallConfigSyncStatusInput{
			ID:                syncID,
			Status:            string(app.StatusError),
			StatusDescription: "failed to enqueue workflow",
		})
		return fmt.Errorf("unable to enqueue execute-workflow signal: %w", err)
	}

	if _, err = callback.AwaitWithTimeout(ctx, cb, callback.FallbackAwaitTimeout); err != nil {
		logger.Error("install sync workflow execution failed", "error", err)
		_ = branchactivities.AwaitUpdateAppInstallConfigSyncStatus(ctx, &branchactivities.UpdateAppInstallConfigSyncStatusInput{
			ID:                syncID,
			Status:            string(app.StatusError),
			StatusDescription: "workflow execution failed",
		})
		return fmt.Errorf("workflow execution failed: %w", err)
	}

	wf, wfErr := workflowactivities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, wfResult.WorkflowID)
	if wfErr == nil && wf.Status.Status != "" {
		switch wf.Status.Status {
		case app.StatusCancelled:
			_ = branchactivities.AwaitUpdateAppInstallConfigSyncStatus(ctx, &branchactivities.UpdateAppInstallConfigSyncStatusInput{
				ID:                syncID,
				Status:            string(app.StatusCancelled),
				StatusDescription: "workflow was cancelled",
			})
			s.updateStepStatus(ctx, app.StatusCancelled, "workflow was cancelled", nil)
			return nil
		case app.StatusError:
			desc := "workflow completed with errors"
			if wf.Status.StatusHumanDescription != "" {
				desc = wf.Status.StatusHumanDescription
			}
			_ = branchactivities.AwaitUpdateAppInstallConfigSyncStatus(ctx, &branchactivities.UpdateAppInstallConfigSyncStatusInput{
				ID:                syncID,
				Status:            string(app.StatusError),
				StatusDescription: desc,
			})
			s.updateStepStatus(ctx, app.StatusError, desc, nil)
			return nil
		}
	}

	_ = branchactivities.AwaitUpdateAppInstallConfigSyncStatus(ctx, &branchactivities.UpdateAppInstallConfigSyncStatusInput{
		ID:                syncID,
		Status:            string(app.StatusSuccess),
		StatusDescription: "sync completed",
	})

	s.updateStepStatus(ctx, app.StatusSuccess, "sync completed", nil)
	logger.Info("install sync completed successfully", "sync_id", syncID)
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
