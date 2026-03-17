package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateAppBranchRunLogStreamInput struct {
	RunID       string `json:"run_id" validate:"required"`
	LogStreamID string `json:"log_stream_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
func (a *Activities) updateAppBranchRunLogStream(ctx context.Context, req *UpdateAppBranchRunLogStreamInput) error {
	if err := a.v.Struct(req); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	if res := a.db.WithContext(ctx).
		Model(&app.AppBranchRun{}).
		Where("id = ?", req.RunID).
		Update("log_stream_id", req.LogStreamID); res.Error != nil {
		return fmt.Errorf("unable to update app branch run log stream: %w", res.Error)
	}

	return nil
}
