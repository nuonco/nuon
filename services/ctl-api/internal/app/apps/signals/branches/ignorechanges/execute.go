package ignorechanges

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	flowdirective "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/directive"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
	workflowactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
	if err != nil {
		return fmt.Errorf("unable to get app branch run: %w", err)
	}
	if run.Force || run.AppBranchConfig.IgnoreChangesRegex == "" {
		return nil
	}
	if run.RunType != app.AppBranchRunTypeGit && run.RunType != app.AppBranchRunTypeGitPreview {
		return nil
	}

	paths := s.ChangedFiles
	if run.AppBranchConfig.IgnoreChangesRegex != ".*" && (run.PRNumber != nil || s.BaseSHA != "" || len(paths) == 0) {
		vcsConfigID := ""
		switch {
		case run.AppBranchConfig.ConnectedGithubVCSConfig != nil:
			vcsConfigID = run.AppBranchConfig.ConnectedGithubVCSConfig.ID
		case run.AppBranchConfig.PublicGitVCSConfig != nil:
			vcsConfigID = run.AppBranchConfig.PublicGitVCSConfig.ID
		}
		if vcsConfigID == "" || run.HeadSHA == "" {
			logger.Warn("unable to evaluate ignored changes; missing VCS config or commit SHA")
			return nil
		}

		changed, fetchErr := activities.AwaitFetchChangedFilePaths(ctx, &activities.FetchChangedFilePathsInput{
			VcsConfigID: vcsConfigID,
			CommitSHA:   run.HeadSHA,
			PRNumber:    run.PRNumber,
			BaseSHA:     s.BaseSHA,
			BaseBranch:  run.BaseBranch,
		})
		if fetchErr != nil {
			logger.Warn("unable to fetch changed files; continuing run", "error", fetchErr)
			return nil
		}
		if changed.Truncated {
			logger.Warn("changed file list was truncated; continuing run")
			return nil
		}
		paths = changed.Paths
	}

	decision, err := Evaluate(run.AppBranchConfig.IgnoreChangesRegex, paths)
	if err != nil {
		logger.Warn("unable to evaluate ignored changes; continuing run", "error", err)
		return nil
	}
	if !decision.Ignored {
		return nil
	}

	if _, err := activities.AwaitUpdateAppBranchRunStatus(ctx, &activities.UpdateAppBranchRunStatusRequest{
		RunID:        run.ID,
		Status:       "not-attempted",
		ErrorMessage: decision.Reason,
	}); err != nil {
		return fmt.Errorf("unable to mark app branch run ignored: %w", err)
	}

	if run.HeadSHA != "" && run.AppBranchConfig.SendStatusesOnIgnore &&
		(!run.IsPreview() || run.PreviewGitHubSetStatuses()) {
		vcsConfigID := ""
		switch {
		case run.AppBranchConfig.ConnectedGithubVCSConfig != nil:
			vcsConfigID = run.AppBranchConfig.ConnectedGithubVCSConfig.ID
		case run.AppBranchConfig.PublicGitVCSConfig != nil:
			vcsConfigID = run.AppBranchConfig.PublicGitVCSConfig.ID
		}
		if vcsConfigID != "" {
			_ = activities.AwaitSetGithubCommitStatus(ctx, &activities.SetGithubCommitStatusInput{
				VcsConfigID: vcsConfigID,
				CommitSHA:   run.HeadSHA,
				State:       "success",
				Description: "Ignored by app branch changes regex",
				AppBranchID: s.AppBranchID,
				RunID:       s.RunID,
				Preview:     run.IsPreview(),
				PreviewMode: run.PreviewMode(),
			})
		}
	}

	if s.StepID != "" {
		if err := statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
			ID: s.StepID,
			Status: app.CompositeStatus{
				Status:                 app.StatusAutoSkipped,
				StatusHumanDescription: decision.Reason,
				Metadata: map[string]any{
					"reason": decision.Reason,
				},
			},
		}); err != nil {
			return fmt.Errorf("unable to update ignore step status: %w", err)
		}
		if err := workflowactivities.AwaitPkgWorkflowsFlowUpdateFlowStepResultDirective(ctx, workflowactivities.UpdateFlowStepResultDirectiveRequest{
			StepID:    s.StepID,
			Directive: string(flowdirective.StepStop),
		}); err != nil {
			return fmt.Errorf("unable to stop ignored app branch run: %w", err)
		}
	}

	return nil
}
