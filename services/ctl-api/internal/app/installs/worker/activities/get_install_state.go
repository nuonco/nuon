package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/types/state"
)

type GetInstallStateRequest struct {
	InstallID string `json:"install_id"`
}

// @temporal-gen activity
// @by-id InstallID
// @timeout 10s
func (a *Activities) GetInstallState(ctx context.Context, req *GetInstallStateRequest) (*state.State, error) {
	state, err := a.helpers.GetInstallState(ctx, req.InstallID, false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install state")
	}

	return state, nil
}
