package activities

import (
	"context"

	"github.com/nuonco/nuon/pkg/types/state"
)

type GetInstallStateRequest struct {
	InstallID string `json:"install_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallID
// @start-to-close-timeout 5m
func (a *Activities) GetInstallState(ctx context.Context, req *GetInstallStateRequest) (*state.State, error) {
	return a.helpers.GetInstallState(ctx, req.InstallID, false, false)
}
