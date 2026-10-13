package activities

import (
	"context"

	pkgstate "github.com/nuonco/nuon/pkg/types/state"
)

type GetLatestInstallStateRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallID
// @start-to-close-timeout 10s
func (a *Activities) GetLatestInstallState(ctx context.Context, req *GetLatestInstallStateRequest) (*pkgstate.State, error) {
	return a.helpers.GetInstallStateBaseline(ctx, req.InstallID)
}
