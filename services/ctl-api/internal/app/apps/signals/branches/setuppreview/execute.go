package setuppreview

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
	if err != nil {
		return fmt.Errorf("unable to get app branch run: %w", err)
	}

	if !run.IsPreview() {
		logger.Info("not a preview run, skipping setup")
		return nil
	}

	// Statuses only need a commit; comments and reactions need a PR.
	needsComment := run.PreviewGitHubComment() && run.PRNumber != nil
	needsStatus := run.PreviewGitHubSetStatuses() && run.HeadSHA != ""
	needsReact := run.PreviewGitHubReact() && run.PRNumber != nil
	if !needsComment && !needsStatus && !needsReact {
		logger.Info("preview GitHub integration disabled, skipping setup")
		return nil
	}

	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	var vcsConfigID string
	if cfg := run.AppBranchConfig.ConnectedGithubVCSConfig; cfg != nil {
		vcsConfigID = cfg.ID
	} else if cfg := run.AppBranchConfig.PublicGitVCSConfig; cfg != nil {
		vcsConfigID = cfg.ID
	}
	if vcsConfigID == "" {
		logger.Info("no VCS config found, skipping GitHub integration")
		return nil
	}

	if run.HeadSHA != "" && run.PreviewGitHubSetStatuses() {
		_ = activities.AwaitSetGithubCommitStatus(ctx, &activities.SetGithubCommitStatusInput{
			VcsConfigID: vcsConfigID,
			CommitSHA:   run.HeadSHA,
			State:       "pending",
			Description: "Preview starting...",
			AppBranchID: s.AppBranchID,
			RunID:       s.RunID,
			Preview:     true,
			PreviewMode: run.PreviewMode(),
		})
	}

	if needsReact {
		if _, err := activities.AwaitCreatePRReaction(ctx, &activities.CreatePRReactionInput{
			VcsConfigID: vcsConfigID,
			PRNumber:    *run.PRNumber,
			Content:     "eyes",
		}); err != nil {
			logger.Warn("unable to add PR reaction", "error", err)
		}
	}

	if !needsComment {
		logger.Info("preview setup complete", "run_id", s.RunID)
		return nil
	}

	prevComment, err := activities.AwaitFindPreviousRunCommentID(ctx, &activities.FindPreviousRunCommentIDInput{
		AppBranchID: s.AppBranchID,
		PRNumber:    *run.PRNumber,
	})
	if err != nil {
		logger.Warn("unable to find previous run comment, will create new", "error", err)
	}

	var existingCommentID *int64
	if prevComment != nil {
		existingCommentID = prevComment.CommentID
	}

	commentContext, _ := activities.AwaitGetPreviewCommentContext(ctx, &activities.GetPreviewCommentContextInput{
		RunID: s.RunID,
	})
	commentBody := activities.BuildPRCommentBody(&activities.PRCommentParams{
		OrgName:     branch.Org.Name,
		AppName:     branch.App.Name,
		AppBranchID: branch.ID,
		BranchName:  branch.Name,
		RunID:       s.RunID,
		HeadSHA:     run.HeadSHA,
		Status:      activities.PRCommentStatusPending,
		Mode:        run.PreviewMode(),
		RunURL:      previewRunURL(commentContext),
	})

	commentResult, err := activities.AwaitCreateOrUpdatePRComment(ctx, &activities.CreateOrUpdatePRCommentInput{
		VcsConfigID:       vcsConfigID,
		PRNumber:          *run.PRNumber,
		AppBranchID:       run.AppBranchID,
		ExistingCommentID: existingCommentID,
		Body:              commentBody,
	})
	if err != nil {
		logger.Warn("unable to post PR comment", "error", err)
		return nil
	}

	if commentResult != nil && commentResult.CommentID != 0 {
		_ = activities.AwaitUpdateAppBranchRunGithubComment(ctx, &activities.UpdateAppBranchRunGithubCommentInput{
			RunID:     s.RunID,
			CommentID: commentResult.CommentID,
		})
		logger.Info("preview comment posted", "comment_id", commentResult.CommentID)
	}

	logger.Info("preview setup complete",
		"run_id", s.RunID,
		"pr_number", *run.PRNumber)

	return nil
}

func previewRunURL(commentContext *activities.GetPreviewCommentContextOutput) string {
	if commentContext == nil {
		return ""
	}
	return commentContext.RunURL
}
