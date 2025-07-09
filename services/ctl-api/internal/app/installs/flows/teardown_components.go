package flows

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
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	steps := make([]*app.WorkflowStep, 0)
	lifecycleSteps, err := getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreTeardownAllComponents)
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

	for _, compID := range componentIDs {
		comp, err := activities.AwaitGetComponentByComponentID(ctx, compID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get component")
		}
		installComp, err := activities.AwaitGetInstallComponent(ctx, activities.GetInstallComponentRequest{
			InstallID:   installID,
			ComponentID: comp.ID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to get install component")
		}

		if installComp == nil {
			continue
		}

		if installComp.StatusV2.Status == app.Status(app.InstallComponentStatusInactive) || installComp.StatusV2.Status == app.Status("") {
			reason := fmt.Sprintf("install component %s is not deployed", comp.Name)
			deployStep, err := installSignalStep(ctx, installID, "skipped teardown "+comp.Name, pgtype.Hstore{
				"reason": &reason,
			}, nil, flw.PlanOnly)
			if err != nil {
				return nil, errors.Wrap(err, "unable to create skip step")
			}
			steps = append(steps, deployStep)

			continue
		}

		preDeploySteps, err := getComponentLifecycleActionsSteps(ctx, flw, compID, installID, app.ActionWorkflowTriggerTypePreTeardownComponent)
		if err != nil {
			return nil, err
		}
		steps = append(steps, preDeploySteps...)

		deployStep, err := installSignalStep(ctx, installID, "teardown "+comp.Name, pgtype.Hstore{}, &signals.Signal{
			Type: signals.OperationExecuteTeardownComponentSyncAndPlan,
			ExecuteTeardownComponentSubSignal: signals.TeardownComponentSubSignal{
				ComponentID: compID,
			},
		}, flw.PlanOnly)
		steps = append(steps, deployStep)

		deployStep, err = installSignalStep(ctx, installID, "teardown "+comp.Name, pgtype.Hstore{}, &signals.Signal{
			Type: signals.OperationExecuteTeardownComponentApplyPlan,
			ExecuteTeardownComponentSubSignal: signals.TeardownComponentSubSignal{
				ComponentID: compID,
			},
		}, flw.PlanOnly)
		steps = append(steps, deployStep)

		postDeploySteps, err := getComponentLifecycleActionsSteps(ctx, flw, compID, installID, app.ActionWorkflowTriggerTypePostTeardownComponent)
		if err != nil {
			return nil, err
		}
		steps = append(steps, postDeploySteps...)
	}

	lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostTeardownAllComponents)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	return steps, nil
}
