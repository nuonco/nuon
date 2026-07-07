package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type FindPreviousRunCommentIDInput struct {
	AppBranchID string `json:"app_branch_id" validate:"required"`
	PRNumber    int    `json:"pr_number" validate:"required"`
}

type FindPreviousRunCommentIDOutput struct {
	CommentID *int64 `json:"comment_id,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) FindPreviousRunCommentID(ctx context.Context, input *FindPreviousRunCommentIDInput) (*FindPreviousRunCommentIDOutput, error) {
	var run app.AppBranchRun
	err := a.db.WithContext(ctx).
		Where(app.AppBranchRun{
			AppBranchID: input.AppBranchID,
			PRNumber:    &input.PRNumber,
		}).
		Where("github_comment_id IS NOT NULL").
		Order("created_at DESC").
		First(&run).Error
	if err != nil {
		return &FindPreviousRunCommentIDOutput{}, nil
	}

	return &FindPreviousRunCommentIDOutput{CommentID: run.GithubCommentID}, nil
}
