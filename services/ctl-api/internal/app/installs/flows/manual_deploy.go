package flows

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.temporal.io/sdk/workflow"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func ManualDeploySteps(ctx workflow.Context, flw *app.Flow) ([]*app.FlowStep, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])

	steps := make([]*app.FlowStep, 0)
	step, err := installSignalStep(ctx, installID, "await runner healthy", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationAwaitRunnerHealthy,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	installDeployID, ok := flw.Metadata["install_deploy_id"]
	if !ok {
		return nil, errors.New("install deploy is not set on the install workflow for a manual deploy")
	}

	deployDependents, _ := flw.Metadata["deploy_dependents"]

	installDeploy, err := activities.AwaitGetDeployByDeployID(ctx, generics.FromPtrStr(installDeployID))
	if err != nil {
		return nil, errors.New("unable to get install deploy")
	}
	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	// first, provision the deploy with before and after triggers
	comp, err := activities.AwaitGetComponentByComponentID(ctx, installDeploy.ComponentID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component")
	}

	preDeploySteps, err := getComponentLifecycleActionsSteps(ctx, flw.ID, installDeploy.ComponentID, installID, app.ActionWorkflowTriggerTypePreDeployComponent)
	if err != nil {
		return nil, err
	}
	steps = append(steps, preDeploySteps...)

	deployStep, err := installSignalStep(ctx, installID, "deploy "+comp.Name, pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationExecuteDeployComponent,
		ExecuteDeployComponentSubSignal: signals.DeployComponentSubSignal{
			DeployID:    generics.FromPtrStr(installDeployID),
			ComponentID: comp.ID,
		},
	})
	steps = append(steps, deployStep)

	postDeploySteps, err := getComponentLifecycleActionsSteps(ctx, flw.ID, installDeploy.ComponentID, installID, app.ActionWorkflowTriggerTypePostDeployComponent)
	if err != nil {
		return nil, err
	}
	steps = append(steps, postDeploySteps...)

	// now queue up any deploy that _depend_ on the input
	componentIDs, err := activities.AwaitGetAppComponentGraph(ctx, activities.GetAppComponentGraphRequest{
		InstallID:   install.ID,
		ComponentID: comp.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app component graph")
	}

	dependencyCompIDs := generics.SliceAfterValue(componentIDs, comp.ID)
	dependencyDeploySteps, err := getComponentDeploySteps(ctx, installID, flw, dependencyCompIDs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component deploy steps")
	}

	if generics.FromPtrStr(deployDependents) == strconv.FormatBool(true) {
		steps = append(steps, dependencyDeploySteps...)
	}

	return steps, nil
}
