package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/types/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
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

	if state.StaleAt != nil {
		a.evClient.Send(ctx, req.InstallID, &signals.Signal{
			Type: signals.OperationGenerateState,
		})
	}

	state, err = a.helpers.GetInstallState(ctx, req.InstallID, false)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install state")
	}

	return state, nil
}
