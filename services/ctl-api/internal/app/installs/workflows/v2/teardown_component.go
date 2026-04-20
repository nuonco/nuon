package v2

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/awaitrunnerhealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/componentteardownapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/componentteardownsyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/generatestate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func TeardownComponent(ctx workflow.Context, flw *app.Workflow) ([]*app.WorkflowStep, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	awData, err := activities.AwaitGetActionWorkflows(ctx, &activities.GetActionWorkflows{
		InstallID: installID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get action workflows")
	}

	sg := newStepGroup()
	steps := make([]*app.WorkflowStep, 0)

	sg.nextGroup() // generate install state
	step, err := sg.installSignalStep(ctx, installID, "generate install state", pgtype.Hstore{}, &generatestate.Signal{
		InstallID: installID,
	}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	componentID, ok := flw.Metadata["component_id"]
	if !ok {
		return nil, errors.New("component id is not set on the install workflow for a manual deploy")
	}

	sg.nextGroup() // await runner health
	step, err = sg.installSignalStep(ctx, installID, "await runner healthy", pgtype.Hstore{}, &awaitrunnerhealthy.Signal{
		InstallID: installID,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	comp, err := activities.AwaitGetComponentByComponentID(ctx, generics.FromPtrStr(componentID))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component")
	}

	preDeploySteps, err := getComponentLifecycleActionsSteps(ctx, flw, comp, installID, app.ActionWorkflowTriggerTypePreTeardownComponent, sg, appCfg, awData)
	if err != nil {
		return nil, err
	}
	steps = append(steps, preDeploySteps...)

	sg.nextGroup() // teardown sync + plan + apply
	if !comp.Type.IsImage() {
		// Resolve install component ID for v2 signals
		installComp, err := activities.AwaitGetInstallComponent(ctx, activities.GetInstallComponentRequest{
			InstallID:   installID,
			ComponentID: comp.ID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to get install component")
		}

		deployStep, err := sg.installSignalStep(ctx, install.ID, "teardown sync and plan "+comp.Name, pgtype.Hstore{}, &componentteardownsyncandplan.Signal{
			InstallComponentID: installComp.ID,
			ComponentID:        generics.FromPtrStr(componentID),
			FlowID:             "",
			SandboxMode:        false,
			Role:               flw.Role,
		}, flw.PlanOnly, WithSkippable(false))
		if err != nil {
			return nil, err
		}
		steps = append(steps, deployStep)

		applyStep, err := sg.installSignalStep(ctx, install.ID, "teardown apply plan "+comp.Name, pgtype.Hstore{}, &componentteardownapplyplan.Signal{
			InstallComponentID: installComp.ID,
			ComponentID:        generics.FromPtrStr(componentID),
			FlowID:             "",
			SandboxMode:        false,
		}, flw.PlanOnly)
		if err != nil {
			return nil, err
		}
		steps = append(steps, applyStep)
	} else {
		deployStep, err := sg.installSignalStep(ctx, installID, "skipped image teardown "+comp.Name, pgtype.Hstore{
			"reason": generics.ToPtr("skipped image teardown"),
		}, nil, false)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create skip step")
		}
		steps = append(steps, deployStep)
	}

	postDeploySteps, err := getComponentLifecycleActionsSteps(ctx, flw, comp, installID, app.ActionWorkflowTriggerTypePostTeardownComponent, sg, appCfg, awData)
	if err != nil {
		return nil, err
	}

	steps = append(steps, postDeploySteps...)

	return steps, nil
}
