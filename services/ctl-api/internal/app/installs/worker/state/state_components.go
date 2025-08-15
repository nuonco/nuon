package state

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/types/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Workflows) getStateComponentsPartial(ctx workflow.Context, installID string) (*state.ComponentsState, error) {
	installComps, err := activities.AwaitGetInstallComponentIDsByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install components")
	}

	st := state.NewComponentsState()
	st.Populated = true

	for _, instCmpID := range installComps {
		compState, err := w.getInstallComponentState(ctx, instCmpID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get install components state")
		}

		st.Components[compState.Name] = compState
	}

	return st, nil
}

func (h *Workflows) getInstallComponentState(ctx workflow.Context, instCompID string) (*state.ComponentState, error) {
	installComp, err := activities.AwaitGetInstallComponentStateByInstallComponentID(ctx, instCompID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install component")
	}

	st := state.NewComponentState()

	st.Name = installComp.Component.Name
	st.Populated = true
	st.ComponentID = installComp.ComponentID
	st.InstallComponentID = installComp.ID

	installDeploys := installComp.InstallDeploys
	if len(installDeploys) < 1 {
		return st, nil
	}

	st.Status = string(installDeploys[0].Status)
	st.BuildID = string(installDeploys[0].ComponentBuildID)
	st.Outputs = installDeploys[0].Outputs

	return st, nil
}
