package workflows

import (
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func TeardownComponents(ctx workflow.Context, flw *app.Workflow) ([]*app.WorkflowStep, error) {
	return teardownComponents(ctx, flw, newStepGroup())
}

//nolint:gocyclo
func teardownComponents(ctx workflow.Context, flw *app.Workflow, sg *stepGroup) ([]*app.WorkflowStep, error) { //nolint:funlen
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	steps := make([]*app.WorkflowStep, 0)

	sg.nextGroup() // generate install state
	step, err := sg.installSignalStep(ctx, installID, "generate install state", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationGenerateState,
	}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}

	steps = append(steps, step)

	lifecycleSteps, err := getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreTeardownAllComponents, sg)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	componentIDs, err := activities.AwaitGetAppGraph(ctx, activities.GetAppGraphRequest{
		InstallID: install.ID,
		Reverse:   true,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install graph")
	}

	appcfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}
	components := make(map[string]app.Component)
	for _, ccc := range appcfg.ComponentConfigConnections {
		components[ccc.ComponentID] = ccc.Component
	}

	for _, compID := range componentIDs {
		sg.nextGroup()
		comp, has := components[compID]
		if !has {
			return nil, errors.Errorf("component %s not found in app config", compID)
		}

		installComp, compErr := activities.AwaitGetInstallComponent(ctx, activities.GetInstallComponentRequest{
			InstallID:   installID,
			ComponentID: comp.ID,
		})
		if compErr != nil {
			return nil, errors.Wrap(compErr, "unable to get install component")
		}

		if installComp == nil {
			continue
		}

		if comp.Type.IsImage() {
			deployStep, imgErr := sg.installSignalStep(ctx, installID, "skipped image teardown "+comp.Name, pgtype.Hstore{
				"reason": generics.ToPtr("skipped image teardown"),
			}, nil, false)
			if imgErr != nil {
				return nil, errors.Wrap(imgErr, "unable to create skip step")
			}

			steps = append(steps, deployStep)
			continue
		}

		if generics.SliceContains(installComp.StatusV2.Status, []app.Status{
			app.Status(app.InstallComponentStatusInactive),
			app.Status(""),
		}) {
			reason := fmt.Sprintf("install component %s is not deployed", comp.Name)

			deployStep, skipErr := sg.installSignalStep(ctx, installID, "skipped teardown "+comp.Name, pgtype.Hstore{
				"reason": generics.ToPtr(reason),
			}, nil, flw.PlanOnly)
			if skipErr != nil {
				return nil, errors.Wrap(skipErr, "unable to create skip step")
			}
			steps = append(steps, deployStep)
			continue
		}

		preDeploySteps, preErr := getComponentLifecycleActionsSteps(ctx, flw, &comp, installID, app.ActionWorkflowTriggerTypePreTeardownComponent, sg)
		if preErr != nil {
			return nil, preErr
		}
		steps = append(steps, preDeploySteps...)

		deployStep, planErr := sg.installSignalStep(ctx, installID, "plan teardown "+comp.Name, pgtype.Hstore{}, &signals.Signal{
			Type: signals.OperationExecuteTeardownComponentSyncAndPlan,
			ExecuteTeardownComponentSubSignal: signals.TeardownComponentSubSignal{
				ComponentID: compID,
			},
		}, flw.PlanOnly, WithSkippable(false))
		_ = planErr
		steps = append(steps, deployStep)

		deployStep, _ = sg.installSignalStep(ctx, installID, "teardown "+comp.Name, pgtype.Hstore{}, &signals.Signal{
			Type: signals.OperationExecuteTeardownComponentApplyPlan,
			ExecuteTeardownComponentSubSignal: signals.TeardownComponentSubSignal{
				ComponentID: compID,
			},
		}, flw.PlanOnly)
		steps = append(steps, deployStep)

		postDeploySteps, postErr := getComponentLifecycleActionsSteps(ctx, flw, &comp, installID, app.ActionWorkflowTriggerTypePostTeardownComponent, sg)
		if postErr != nil {
			return nil, postErr
		}
		steps = append(steps, postDeploySteps...)
	}

	lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostTeardownAllComponents, sg)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	return steps, nil
}
