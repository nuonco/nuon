package flows

import (
	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.temporal.io/sdk/workflow"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func TeardownComponent(ctx workflow.Context, flw *app.Flow) ([]*app.FlowStep, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	componentID, ok := flw.Metadata["component_id"]
	if !ok {
		return nil, errors.New("component id is not set on the install workflow for a manual deploy")
	}

	steps := make([]*app.FlowStep, 0)

	comp, err := activities.AwaitGetComponentByComponentID(ctx, generics.FromPtrStr(componentID))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component")
	}

	preDeploySteps, err := getComponentLifecycleActionsSteps(ctx, flw.ID, generics.FromPtrStr(componentID), installID, app.ActionWorkflowTriggerTypePreTeardownComponent)
	if err != nil {
		return nil, err
	}
	steps = append(steps, preDeploySteps...)

	deployStep, err := installSignalStep(ctx, install.ID, "teardown "+comp.Name, pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationExecuteTeardownComponent,
		ExecuteTeardownComponentSubSignal: signals.TeardownComponentSubSignal{
			ComponentID: generics.FromPtrStr(componentID),
		},
	})
	steps = append(steps, deployStep)

	postDeploySteps, err := getComponentLifecycleActionsSteps(ctx, flw.ID, generics.FromPtrStr(componentID), installID, app.ActionWorkflowTriggerTypePostTeardownComponent)
	if err != nil {
		return nil, err
	}

	steps = append(steps, postDeploySteps...)

	return steps, nil
}
