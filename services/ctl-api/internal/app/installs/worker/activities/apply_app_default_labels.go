package activities

import (
	"context"
)

type ApplyAppDefaultLabelsRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) ApplyAppDefaultLabels(ctx context.Context, req *ApplyAppDefaultLabelsRequest) error {
	return a.helpers.ApplyAppDefaultLabels(ctx, req.InstallID)
}
