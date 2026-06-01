package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CloseLogStreamRequest struct {
	LogStreamID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CloseLogStream(ctx context.Context, req CloseLogStreamRequest) error {
	if err := a.v.Struct(req); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	if res := a.db.WithContext(ctx).Model(&app.LogStream{}).
		Where("id = ?", req.LogStreamID).
		Update("open", false); res.Error != nil {
		return fmt.Errorf("unable to close log stream: %w", res.Error)
	}
	return nil
}
