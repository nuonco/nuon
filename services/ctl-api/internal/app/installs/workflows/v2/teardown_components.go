package v2

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentteardownapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentteardownsyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generatestate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

func TeardownComponents(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
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

	sg := newStepGroup(flw)
	dg := newGenCtx(sg, flw, installID, appCfg, awData)
	steps, err := teardownComponents(ctx, dg, install)
	if err != nil {
		return nil, err
	}
	return sg.Result(steps), nil
}

func teardownComponents(ctx workflow.Context, dg *genCtx, install *app.Install) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0)

	dg.sg.nextGroupEager() // generate install state
	orgEnabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureStateGenV2))
	if err != nil {
		return nil, errors.Wrap(err, "unable to check state-gen-v2 feature")
	}
	stateGenV2 := statemanager.UseStateGenV2(orgEnabled, install.Metadata)

	if !stateGenV2 {
		stateSignal := &generatestate.Signal{InstallID: dg.installID}
		step, err := dg.sg.installSignalStep(ctx, dg.installID, "generate install state", pgtype.Hstore{}, stateSignal, dg.flw.PlanOnly, WithSkippable(false))
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	lifecycleSteps, err := getLifecycleActionsSteps(ctx, dg, app.ActionWorkflowTriggerTypePreTeardownAllComponents)
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
		dg.sg.nextGroup() // new group for each component
		comp, has := dg.components[compID]
		if !has {
			return nil, errors.Errorf("component %s not found in app config", compID)
		}

		installComp, err := activities.AwaitGetInstallComponent(ctx, activities.GetInstallComponentRequest{
			InstallID:   dg.installID,
			ComponentID: comp.ID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to get install component")
		}

		if installComp == nil {
			continue
		}

		if comp.Type.IsImage() {
			deployStep, err := dg.sg.installSignalStep(ctx, dg.installID, "skipped image teardown "+comp.Name, pgtype.Hstore{
				"reason": generics.ToPtr("skipped image teardown"),
			}, nil, false)
			if err != nil {
				return nil, errors.Wrap(err, "unable to create skip step")
			}

			steps = append(steps, deployStep)
			continue
		}

		if generics.SliceContains(installComp.StatusV2.Status, []app.Status{
			app.Status(app.InstallComponentStatusInactive),
			app.Status(""),
		}) {
			reason := fmt.Sprintf("install component %s is not deployed", comp.Name)

			deployStep, err := dg.sg.installSignalStep(ctx, dg.installID, "skipped teardown "+comp.Name, pgtype.Hstore{
				"reason": generics.ToPtr(reason),
			}, nil, dg.flw.PlanOnly)
			if err != nil {
				return nil, errors.Wrap(err, "unable to create skip step")
			}
			steps = append(steps, deployStep)
			continue
		}

		preDeploySteps, err := getComponentLifecycleActionsSteps(ctx, dg, &comp, app.ActionWorkflowTriggerTypePreTeardownComponent)
		if err != nil {
			return nil, err
		}
		steps = append(steps, preDeploySteps...)

		deployStep, err := dg.sg.installSignalStep(ctx, dg.installID, "plan teardown "+comp.Name, pgtype.Hstore{}, &componentteardownsyncandplan.Signal{
			InstallComponentID: installComp.ID,
			InstallID:          dg.installID,
			ComponentID:        compID,
			FlowID:             "",
			SandboxMode:        false,
			Role:               dg.flw.Role,
		}, dg.flw.PlanOnly, WithSkippable(false))
		if err != nil {
			return nil, err
		}
		steps = append(steps, deployStep)

		deployStep, err = dg.sg.installSignalStep(ctx, dg.installID, "teardown "+comp.Name, pgtype.Hstore{}, &componentteardownapplyplan.Signal{
			InstallComponentID: installComp.ID,
			InstallID:          dg.installID,
			ComponentID:        compID,
			FlowID:             "",
			SandboxMode:        false,
		}, dg.flw.PlanOnly, WithMaxAutoRetries(componentMaxAutoRetries(dg.appCfg, compID)))
		if err != nil {
			return nil, err
		}
		steps = append(steps, deployStep)

		postDeploySteps, err := getComponentLifecycleActionsSteps(ctx, dg, &comp, app.ActionWorkflowTriggerTypePostTeardownComponent)
		if err != nil {
			return nil, err
		}
		steps = append(steps, postDeploySteps...)
	}

	lifecycleSteps, err = getLifecycleActionsSteps(ctx, dg, app.ActionWorkflowTriggerTypePostTeardownAllComponents)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	return steps, nil
}
