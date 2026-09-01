package run

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

const previewCommitStatusesVersion = "app-branch-preview-commit-statuses-v1"

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	previewStatusesEnabled := workflow.GetVersion(ctx, previewCommitStatusesVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion

	// Fetch run from DB
	run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
	if err != nil {
		return fmt.Errorf("unable to get app branch run: %w", err)
	}

	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, run.AppBranchID)
	if err != nil {
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	logger.Info("starting app branch run",
		"run_id", run.ID,
		"app_branch_id", branch.ID,
		"app_branch_name", branch.Name,
		"workflow_id", *run.WorkflowID,
		"force", run.Force,
	)

	// Update status to running
	if _, err = activities.AwaitUpdateAppBranchRunStatus(ctx, &activities.UpdateAppBranchRunStatusRequest{
		RunID:  run.ID,
		Status: "running",
	}); err != nil {
		logger.Error("unable to update run status to running", "error", err)
		return err
	}

	// Enqueue the shared execute-workflow signal to the branch's queue
	cb := callback.New(ctx, run.ID)
	enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:         branch.ID,
		OwnerType:       "app_branches",
		SignalOwnerID:   *run.WorkflowID,
		SignalOwnerType: "install_workflows",
		Signal: &executeflow.Signal{
			WorkflowID: *run.WorkflowID,
		},
		Callback: cb,
	})
	if err != nil {
		logger.Error("unable to enqueue execute-workflow signal", "error", err)
		if previewStatusesEnabled {
			s.setPreviewCommitStatus(ctx, run, "failure", "Preview failed")
		}
		if _, updateErr := activities.AwaitUpdateAppBranchRunStatus(ctx, &activities.UpdateAppBranchRunStatusRequest{
			RunID:        run.ID,
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("enqueue failed: %v", err),
		}); updateErr != nil {
			logger.Error("unable to update run status to failed", "error", updateErr)
		}
		return fmt.Errorf("unable to enqueue execute-workflow signal: %w", err)
	}

	logger.Info("waiting for workflow execution to complete",
		"queue_signal_id", enqueueResp.QueueSignalID,
	)

	// Await the execute-workflow signal completion
	if _, err = callback.AwaitWithTimeout(ctx, cb, callback.FallbackAwaitTimeout); err != nil {
		if previewStatusesEnabled {
			latestRun, getErr := activities.AwaitGetAppBranchRunByIDByRunID(ctx, run.ID)
			if getErr == nil && latestRun.Status == "not-attempted" {
				_ = statusactivities.AwaitPkgStatusUpdateFlowStatus(ctx, statusactivities.UpdateStatusRequest{
					ID: *run.WorkflowID,
					Status: app.CompositeStatus{
						Status:                 app.StatusNotAttempted,
						StatusHumanDescription: latestRun.ErrorMessage,
						Metadata: map[string]any{
							"ignored_by_regex": true,
						},
					},
				})
				return nil
			}
		}

		logger.Error("workflow execution failed", "error", err)
		if previewStatusesEnabled {
			s.setPreviewCommitStatus(ctx, run, "failure", "Preview failed")
		}
		if _, updateErr := activities.AwaitUpdateAppBranchRunStatus(ctx, &activities.UpdateAppBranchRunStatusRequest{
			RunID:        run.ID,
			Status:       "failed",
			ErrorMessage: fmt.Sprintf("workflow execution failed: %v", err),
		}); updateErr != nil {
			logger.Error("unable to update run status to failed", "error", updateErr)
		}
		return fmt.Errorf("workflow execution failed: %w", err)
	}

	// Re-fetch the workflow to check actual status — the callback completing
	// doesn't mean all steps succeeded (cancelled/errored steps are terminal).
	wf, wfErr := workflowactivities.AwaitPkgWorkflowsFlowGetFlowByID(ctx, *run.WorkflowID)
	if wfErr == nil && wf.Status.Status != "" {
		switch wf.Status.Status {
		case app.StatusCancelled:
			if previewStatusesEnabled {
				s.setPreviewCommitStatus(ctx, run, "failure", "Preview cancelled")
			}
			if _, updateErr := activities.AwaitUpdateAppBranchRunStatus(ctx, &activities.UpdateAppBranchRunStatusRequest{
				RunID:        run.ID,
				Status:       "cancelled",
				ErrorMessage: "workflow was cancelled",
			}); updateErr != nil {
				logger.Error("unable to update run status to cancelled", "error", updateErr)
			}
			return nil
		case app.StatusError:
			if previewStatusesEnabled {
				s.setPreviewCommitStatus(ctx, run, "failure", "Preview failed")
			}
			errMsg := "workflow completed with errors"
			if wf.Status.StatusHumanDescription != "" {
				errMsg = wf.Status.StatusHumanDescription
			}
			if _, updateErr := activities.AwaitUpdateAppBranchRunStatus(ctx, &activities.UpdateAppBranchRunStatusRequest{
				RunID:        run.ID,
				Status:       "failed",
				ErrorMessage: errMsg,
			}); updateErr != nil {
				logger.Error("unable to update run status to failed", "error", updateErr)
			}
			return nil
		}
	}

	if previewStatusesEnabled {
		s.setPreviewCommitStatus(ctx, run, "success", "Preview complete")
	}
	if _, err = activities.AwaitUpdateAppBranchRunStatus(ctx, &activities.UpdateAppBranchRunStatusRequest{
		RunID:  run.ID,
		Status: "success",
	}); err != nil {
		logger.Error("unable to update run status to success", "error", err)
	}

	logger.Info("app branch run completed successfully",
		"run_id", run.ID,
		"app_branch_id", branch.ID,
		"workflow_id", *run.WorkflowID,
	)

	return nil
}

func (s *Signal) setPreviewCommitStatus(ctx workflow.Context, run *app.AppBranchRun, state, description string) {
	if !run.IsPreview() || !run.PreviewGitHubSetStatuses() || run.HeadSHA == "" {
		return
	}

	vcsConfigID := ""
	switch {
	case run.AppBranchConfig.ConnectedGithubVCSConfig != nil:
		vcsConfigID = run.AppBranchConfig.ConnectedGithubVCSConfig.ID
	case run.AppBranchConfig.PublicGitVCSConfig != nil:
		vcsConfigID = run.AppBranchConfig.PublicGitVCSConfig.ID
	}
	if vcsConfigID == "" {
		return
	}

	_ = activities.AwaitSetGithubCommitStatus(ctx, &activities.SetGithubCommitStatusInput{
		VcsConfigID: vcsConfigID,
		CommitSHA:   run.HeadSHA,
		State:       state,
		Description: description,
		AppBranchID: run.AppBranchID,
		RunID:       run.ID,
	})
}
