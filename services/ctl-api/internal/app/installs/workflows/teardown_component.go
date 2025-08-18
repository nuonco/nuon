package workflows

import (
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func TeardownComponent(ctx workflow.Context, flw *app.Workflow) ([]*app.WorkflowStep, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	sg := newStepGroup()

	sg.nextGroup() // generate install state
	step, err := sg.installSignalStep(ctx, installID, "generate install state", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationGenerateState,
	}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}

	componentID, ok := flw.Metadata["component_id"]
	if !ok {
		return nil, errors.New("component id is not set on the install workflow for a manual deploy")
	}

	steps := make([]*app.WorkflowStep, 0)
	sg.nextGroup() // await runner health
	step, err = sg.installSignalStep(ctx, installID, "await runner healthy", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationAwaitRunnerHealthy,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	comp, err := activities.AwaitGetComponentByComponentID(ctx, generics.FromPtrStr(componentID))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component")
	}

	preDeploySteps, err := getComponentLifecycleActionsSteps(ctx, flw, comp, installID, app.ActionWorkflowTriggerTypePreTeardownComponent, sg)
	if err != nil {
		return nil, err
	}
	steps = append(steps, preDeploySteps...)

	sg.nextGroup() // teardown sync + plan + apply
	deployStep, err := sg.installSignalStep(ctx, install.ID, "teardown sync and plan "+comp.Name, pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationExecuteTeardownComponentSyncAndPlan,
		ExecuteTeardownComponentSubSignal: signals.TeardownComponentSubSignal{
			ComponentID: generics.FromPtrStr(componentID),
		},
	}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}
	steps = append(steps, deployStep)

	applyStep, err := sg.installSignalStep(ctx, install.ID, "teardown apply plan "+comp.Name, pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationExecuteTeardownComponentApplyPlan,
		ExecuteTeardownComponentSubSignal: signals.TeardownComponentSubSignal{
			ComponentID: generics.FromPtrStr(componentID),
		},
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, applyStep)

	postDeploySteps, err := getComponentLifecycleActionsSteps(ctx, flw, comp, installID, app.ActionWorkflowTriggerTypePostTeardownComponent, sg)
	if err != nil {
		return nil, err
	}

	steps = append(steps, postDeploySteps...)

	return steps, nil
}
