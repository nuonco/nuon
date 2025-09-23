package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetComponentInstallsRequest struct {
	ComponentID string `validate:"required"`
	AppID       string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) GetComponentInstalls(ctx context.Context, req GetComponentInstallsRequest) ([]string, error) {
	installs, err := a.appsHelpers.GetAppInstalls(ctx, req.AppID)
	if err != nil {
		return nil, fmt.Errorf("unable to get component dependents: %w", err)
	}

	activeInstalls := make([]string, 0)
	for _, inst := range installs {
		// if an install was never attempted, it does not need to be polled
		if len(inst.InstallSandboxRuns) < 1 {
			continue
		}

		if inst.InstallSandboxRuns[0].Status == app.SandboxRunStatusAccessError ||
			inst.InstallSandboxRuns[0].Status == app.SandboxRunStatusDeprovisioned {
			continue
		}

		for _, instComp := range inst.InstallComponents {
			if instComp.ComponentID == req.ComponentID {
				if instComp.Status == app.InstallComponentStatusInactive || instComp.Status == "" {
					continue
				}
				activeInstalls = append(activeInstalls, inst.ID)
				break
			}
		}
	}

	return activeInstalls, nil
}
