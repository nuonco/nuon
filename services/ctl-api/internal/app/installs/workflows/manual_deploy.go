package workflows

import (
	"strconv"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

//nolint:gocyclo
func ManualDeploySteps(ctx workflow.Context, flw *app.Workflow) ([]*app.WorkflowStep, error) { //nolint:funlen
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	sg := newStepGroup()

	steps := make([]*app.WorkflowStep, 0)

	sg.nextGroup() // generate install state
	step, err := sg.installSignalStep(ctx, installID, "generate install state", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationGenerateState,
	}, flw.PlanOnly, WithSkippable(false))
	_ = err
	steps = append(steps, step)

	sg.nextGroup() // runner health
	step, err = sg.installSignalStep(ctx, installID, "await runner healthy", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationAwaitRunnerHealthy,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	installDeployID, ok := flw.Metadata["install_deploy_id"]
	if !ok {
		return nil, errors.New("install deploy is not set on the install workflow for a manual deploy")
	}

	deployDependents := flw.Metadata["deploy_dependents"]

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
	preDeploySteps, err := getComponentLifecycleActionsSteps(
		ctx,
		flw,
		comp,
		installID,
		app.ActionWorkflowTriggerTypePreDeployComponent,
		sg,
	)
	if err != nil {
		return nil, err
	}
	if !flw.PlanOnly {
		steps = append(steps, preDeploySteps...)
	}

	// sync image
	if comp.Type.IsImage() {
		sg.nextGroup() // component sync
		deployStep, syncErr := sg.installSignalStep(ctx, installID, "sync "+comp.Name, pgtype.Hstore{}, &signals.Signal{
			Type: signals.OperationExecuteDeployComponentSyncImage,
			ExecuteDeployComponentSubSignal: signals.DeployComponentSubSignal{
				DeployID:    generics.FromPtrStr(installDeployID),
				ComponentID: comp.ID,
			},
		}, flw.PlanOnly)
		if syncErr != nil {
			return nil, errors.Wrap(syncErr, "unable to create image sync")
		}

		steps = append(steps, deployStep)
	} else {
		sg.nextGroup() // component sync + plan + apply
		planStep, planErr := sg.installSignalStep(ctx, installID, "sync and plan "+comp.Name, pgtype.Hstore{}, &signals.Signal{
			Type: signals.OperationExecuteDeployComponentSyncAndPlan,
			ExecuteDeployComponentSubSignal: signals.DeployComponentSubSignal{
				DeployID:    generics.FromPtrStr(installDeployID),
				ComponentID: comp.ID,
			},
		}, flw.PlanOnly, WithSkippable(false))
		if planErr != nil {
			return nil, errors.Wrap(planErr, "unable to create image sync")
		}
		applyPlanStep, applyErr := sg.installSignalStep(ctx, installID, "apply "+comp.Name, pgtype.Hstore{}, &signals.Signal{
			Type: signals.OperationExecuteDeployComponentApplyPlan,
			ExecuteDeployComponentSubSignal: signals.DeployComponentSubSignal{
				DeployID:    generics.FromPtrStr(installDeployID),
				ComponentID: comp.ID,
			},
		}, flw.PlanOnly)
		if applyErr != nil {
			return nil, errors.Wrap(applyErr, "unable to create image sync")
		}

		if flw.PlanOnly {
			steps = append(steps, planStep)
		} else {
			steps = append(steps, planStep, applyPlanStep)
		}
	}

	postDeploySteps, err := getComponentLifecycleActionsSteps(ctx, flw, comp, installID, app.ActionWorkflowTriggerTypePostDeployComponent, sg)
	if err != nil {
		return nil, err
	}
	if !flw.PlanOnly {
		steps = append(steps, postDeploySteps...)
	}

	// now queue up any deploy that _depend_ on the input
	componentIDs, err := activities.AwaitGetAppComponentGraph(ctx, activities.GetAppComponentGraphRequest{
		InstallID:   install.ID,
		ComponentID: comp.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app component graph")
	}

	dependencyCompIDs := generics.SliceAfterValue(componentIDs, comp.ID)
	dependencyDeploySteps, err := getComponentDeploySteps(ctx, installID, flw, dependencyCompIDs, sg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component deploy steps")
	}

	if generics.FromPtrStr(deployDependents) == strconv.FormatBool(true) {
		steps = append(steps, dependencyDeploySteps...)
	}

	return steps, nil
}
