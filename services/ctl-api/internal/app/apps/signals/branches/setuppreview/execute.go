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

	if run.PRNumber == nil {
		logger.Info("no PR number on preview run, skipping GitHub integration")
		return nil
	}

	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	var vcsConfigID string
	if len(branch.Configs) > 0 {
		if cfg := branch.Configs[0].ConnectedGithubVCSConfig; cfg != nil {
			vcsConfigID = cfg.ID
		} else if cfg := branch.Configs[0].PublicGitVCSConfig; cfg != nil {
			vcsConfigID = cfg.ID
		}
	}
	if vcsConfigID == "" {
		logger.Info("no VCS config found, skipping GitHub integration")
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

	commentBody := activities.BuildPRCommentBody(&activities.PRCommentParams{
		AppName: branch.Name,
		RunID:   s.RunID,
		Status:  activities.PRCommentStatusPending,
	})

	if run.HeadSHA != "" {
		_ = activities.AwaitSetGithubCommitStatus(ctx, &activities.SetGithubCommitStatusInput{
			VcsConfigID: vcsConfigID,
			CommitSHA:   run.HeadSHA,
			State:       "pending",
			Context:     "nuon/preview",
			Description: "Preview starting...",
		})
	}

	commentResult, err := activities.AwaitCreateOrUpdatePRComment(ctx, &activities.CreateOrUpdatePRCommentInput{
		VcsConfigID:       vcsConfigID,
		PRNumber:          *run.PRNumber,
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
	}

	logger.Info("preview setup complete",
		"run_id", s.RunID,
		"pr_number", *run.PRNumber,
		"comment_id", commentResult.CommentID)

	return nil
}
