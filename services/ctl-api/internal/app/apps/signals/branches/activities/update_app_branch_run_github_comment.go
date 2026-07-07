package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateAppBranchRunGithubCommentInput struct {
	RunID     string `json:"run_id" validate:"required"`
	CommentID int64  `json:"comment_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) UpdateAppBranchRunGithubComment(ctx context.Context, input *UpdateAppBranchRunGithubCommentInput) error {
	res := a.db.WithContext(ctx).
		Model(&app.AppBranchRun{}).
		Where(app.AppBranchRun{ID: input.RunID}).
		Update("github_comment_id", input.CommentID)
	if res.Error != nil {
		return fmt.Errorf("unable to update github comment id: %w", res.Error)
	}

	return nil
}
