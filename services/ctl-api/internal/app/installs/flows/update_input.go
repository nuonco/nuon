package flows

import (
	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func InputUpdate(ctx workflow.Context, flw *app.Flow) ([]*app.FlowStep, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}
	steps := make([]*app.InstallWorkflowStep, 0)
	lifecycleSteps, err := getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreUpdateInputs)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	componentIDs, err := activities.AwaitGetAppGraph(ctx, activities.GetAppGraphRequest{
		InstallID: install.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install graph")
	}

	deploySteps, err := getComponentDeploySteps(ctx, installID, flw, componentIDs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component deploy steps")
	}
	steps = append(steps, deploySteps...)

	lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostUpdateInputs)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	return steps, nil
}
