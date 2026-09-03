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

const (
	previewCommitStatusesVersion      = "app-branch-preview-commit-statuses-v1"
	previewCommentRunFinalizerVersion = "app-branch-preview-comment-run-finalizer-v1"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	previewStatusesEnabled := workflow.GetVersion(ctx, previewCommitStatusesVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion
	previewCommentFinalizerEnabled := workflow.GetVersion(ctx, previewCommentRunFinalizerVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion

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
		if previewCommentFinalizerEnabled {
			s.writeFinalizerComment(ctx, run, branch, activities.PRCommentStatusFailed, fmt.Sprintf("enqueue failed: %v", err))
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
		if previewCommentFinalizerEnabled {
			s.writeFinalizerComment(ctx, run, branch, activities.PRCommentStatusFailed, fmt.Sprintf("workflow execution failed: %v", err))
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
			if previewCommentFinalizerEnabled {
				s.writeFinalizerComment(ctx, run, branch, activities.PRCommentStatusFailed, "workflow was cancelled")
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
			if previewCommentFinalizerEnabled {
				s.writeFinalizerComment(ctx, run, branch, activities.PRCommentStatusFailed, errMsg)
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
	if previewCommentFinalizerEnabled {
		s.writeFinalizerComment(ctx, run, branch, activities.PRCommentStatusSuccess, "")
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
		Preview:     true,
		PreviewMode: run.PreviewMode(),
	})
}

func (s *Signal) writeFinalizerComment(ctx workflow.Context, run *app.AppBranchRun, branch *app.AppBranch, status activities.PRCommentStatus, errorMsg string) {
	if run.PRNumber == nil || !run.PreviewGitHubComment() {
		return
	}

	var vcsConfigID string
	switch {
	case run.AppBranchConfig.ConnectedGithubVCSConfig != nil:
		vcsConfigID = run.AppBranchConfig.ConnectedGithubVCSConfig.ID
	case run.AppBranchConfig.PublicGitVCSConfig != nil:
		vcsConfigID = run.AppBranchConfig.PublicGitVCSConfig.ID
	}
	if vcsConfigID == "" {
		return
	}

	commentContext, _ := activities.AwaitGetPreviewCommentContext(ctx, &activities.GetPreviewCommentContextInput{
		RunID: s.RunID,
	})

	phases := finalizerPhases(commentContext, run.PreviewMode(), status == activities.PRCommentStatusSuccess)

	installApplied := status == activities.PRCommentStatusSuccess && run.PreviewMode() == app.AppBranchRunPreviewModeApply
	var previewInstallName, previewInstallURL string
	if installApplied && commentContext != nil {
		previewInstallName = commentContext.PreviewInstallName
		previewInstallURL = commentContext.PreviewInstallURL
	}

	commentBody := activities.BuildPRCommentBody(&activities.PRCommentParams{
		OrgName:            branch.Org.Name,
		AppName:            branch.App.Name,
		BranchName:         branch.Name,
		RunID:              s.RunID,
		RunURL:             previewRunURL(commentContext),
		Status:             status,
		Mode:               run.PreviewMode(),
		Diff:               previewDiff(commentContext),
		ComponentChanges:   previewComponentChanges(commentContext),
		InstallImpact:      previewInstallImpact(commentContext),
		PreviewInstallName: previewInstallName,
		PreviewInstallURL:  previewInstallURL,
		InstallApplied:     installApplied,
		ErrorMessage:       errorMsg,
		Phases:             phases,
	})

	_, _ = activities.AwaitCreateOrUpdatePRComment(ctx, &activities.CreateOrUpdatePRCommentInput{
		VcsConfigID:       vcsConfigID,
		PRNumber:          *run.PRNumber,
		ExistingCommentID: run.GithubCommentID,
		Body:              commentBody,
	})
}

func finalizerPhases(commentContext *activities.GetPreviewCommentContextOutput, mode app.AppBranchRunPreviewMode, success bool) *activities.PRCommentPhases {
	phases := commentContextPhases(commentContext)
	if success {
		phases.Config = activities.PRCommentPhaseValid
		phases.Builds = activities.PRCommentPhaseValid
		if mode != app.AppBranchRunPreviewModeBuildOnly {
			phases.Install = activities.PRCommentPhaseValid
		} else {
			phases.Install = ""
		}
	} else {
		activities.FinalizeFailedPhases(phases)
	}
	return phases
}

func commentContextPhases(commentContext *activities.GetPreviewCommentContextOutput) *activities.PRCommentPhases {
	if commentContext == nil || commentContext.Phases == nil {
		return &activities.PRCommentPhases{}
	}
	cp := *commentContext.Phases
	return &cp
}

func previewRunURL(commentContext *activities.GetPreviewCommentContextOutput) string {
	if commentContext == nil {
		return ""
	}
	return commentContext.RunURL
}

func previewComponentChanges(commentContext *activities.GetPreviewCommentContextOutput) []activities.ComponentBuildChange {
	if commentContext == nil {
		return nil
	}
	return commentContext.ComponentChanges
}

func previewDiff(commentContext *activities.GetPreviewCommentContextOutput) *activities.ComputeAppConfigDiffOutput {
	if commentContext == nil {
		return nil
	}
	return commentContext.Diff
}

func previewInstallImpact(commentContext *activities.GetPreviewCommentContextOutput) []activities.InstallGroupImpact {
	if commentContext == nil {
		return nil
	}
	return commentContext.InstallImpact
}
