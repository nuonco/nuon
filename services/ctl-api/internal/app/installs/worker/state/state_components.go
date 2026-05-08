package state

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

func (w *Workflows) getStateComponentsPartial(ctx workflow.Context, installID string) (*state.ComponentsState, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get logger")
	}

	installComps, err := activities.AwaitGetInstallComponentIDsByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install components")
	}

	l.Info("legacy state: fetching component states", zap.Int("component_count", len(installComps)))

	st := state.NewComponentsState()
	st.Populated = true

	for _, instCmpID := range installComps {
		compState, err := w.getInstallComponentState(ctx, instCmpID, l)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get install components state")
		}

		st.Components[compState.Name] = compState
	}

	return st, nil
}

func (h *Workflows) getInstallComponentState(ctx workflow.Context, instCompID string, l *zap.Logger) (*state.ComponentState, error) {
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

	l.Info("legacy state: component outputs",
		zap.String("component_name", st.Name),
		zap.String("install_component_id", instCompID),
		zap.Int("output_count", len(st.Outputs)),
		zap.Bool("has_outputs", len(st.Outputs) > 0),
	)

	return st, nil
}
