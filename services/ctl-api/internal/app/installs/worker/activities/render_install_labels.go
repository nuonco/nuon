package activities

import (
	"context"
)

type RenderInstallLabelsRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) RenderInstallLabels(ctx context.Context, req *RenderInstallLabelsRequest) error {
	return a.helpers.RenderInstallLabels(ctx, req.InstallID)
}
