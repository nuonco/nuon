package v2

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitrunnerhealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentdeployapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentdeploysyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentsyncimage"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generatestate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

func ComponentEnabledSteps(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
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
	dg := newGenCtx(sg, flw, installID, appCfg, awData, WithInstallConfig(install.InstallConfig))
	steps := make([]*app.WorkflowStep, 0)

	sg.nextGroupEager()
	orgEnabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureStateGenV2))
	if err != nil {
		return nil, errors.Wrap(err, "unable to check state-gen-v2 feature")
	}
	stateGenV2 := statemanager.UseStateGenV2(orgEnabled, install.Metadata)

	if !stateGenV2 {
		stateSignal := &generatestate.Signal{InstallID: installID}
		step, err := sg.installSignalStep(ctx, installID, "generate install state", pgtype.Hstore{}, stateSignal, flw.PlanOnly, WithSkippable(false))
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	componentID, ok := flw.Metadata["component_id"]
	if !ok {
		return nil, errors.New("component_id is not set on the workflow metadata")
	}

	sg.nextGroupEager()
	step, err := sg.installSignalStep(ctx, installID, "runner healthy", pgtype.Hstore{}, &awaitrunnerhealthy.Signal{
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

	installComp, err := activities.AwaitGetInstallComponent(ctx, activities.GetInstallComponentRequest{
		InstallID:   installID,
		ComponentID: comp.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install component")
	}

	preEnableSteps, err := getComponentLifecycleActionsSteps(ctx, dg, comp, app.ActionWorkflowTriggerTypePreEnableComponent)
	if err != nil {
		return nil, err
	}
	if !flw.PlanOnly {
		steps = append(steps, preEnableSteps...)
	}

	compIDStr := generics.FromPtrStr(componentID)

	if comp.Type.IsImage() {
		sg.nextGroup()
		latestBuild, err := activities.AwaitGetLatestActiveComponentBuildByComponentID(ctx, comp.ID)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to resolve latest active build for image component %s", comp.Name)
		}
		var buildID string
		if latestBuild != nil {
			buildID = latestBuild.ID
		}
		deployStep, err := sg.installSignalStep(ctx, installID, "sync "+comp.Name, pgtype.Hstore{}, &componentsyncimage.Signal{
			InstallComponentID: installComp.ID,
			ComponentID:        comp.ID,
			BuildID:            buildID,
			Role:               flw.Role,
		}, flw.PlanOnly)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create image sync")
		}
		steps = append(steps, deployStep)
	} else {
		if !flw.PlanOnly && !comp.Type.IsImage() {
			depSyncSteps, err := getImageDepSyncSteps(ctx, dg, comp.ID, 0, map[string]int{comp.ID: 0})
			if err != nil {
				return nil, errors.Wrap(err, "unable to prepend image-dep sync steps")
			}
			if len(depSyncSteps) > 0 {
				steps = append(steps, depSyncSteps...)
				sg.nextGroup()
			}
		}

		sg.nextGroup()
		latestBuild, err := activities.AwaitGetLatestActiveComponentBuildByComponentID(ctx, comp.ID)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to resolve latest active build for component %s", comp.Name)
		}
		var buildID string
		if latestBuild != nil {
			buildID = latestBuild.ID
		}
		planStep, err := sg.installSignalStep(ctx, installID, "sync and plan "+comp.Name, pgtype.Hstore{}, &componentdeploysyncandplan.Signal{
			InstallComponentID: installComp.ID,
			InstallID:          installID,
			ComponentID:        compIDStr,
			BuildID:            buildID,
			Role:               flw.Role,
		}, flw.PlanOnly, WithSkippable(false))
		if err != nil {
			return nil, errors.Wrap(err, "unable to create sync and plan step")
		}

		applyStep, err := sg.installSignalStep(ctx, installID, "apply "+comp.Name, pgtype.Hstore{}, &componentdeployapplyplan.Signal{
			InstallComponentID: installComp.ID,
			InstallID:          installID,
			ComponentID:        compIDStr,
		}, flw.PlanOnly, WithMaxAutoRetries(componentMaxAutoRetries(appCfg, compIDStr)))
		if err != nil {
			return nil, errors.Wrap(err, "unable to create apply step")
		}

		if flw.PlanOnly {
			steps = append(steps, planStep)
		} else {
			steps = append(steps, planStep, applyStep)
		}
	}

	postEnableSteps, err := getComponentLifecycleActionsSteps(ctx, dg, comp, app.ActionWorkflowTriggerTypePostEnableComponent)
	if err != nil {
		return nil, err
	}
	if !flw.PlanOnly {
		steps = append(steps, postEnableSteps...)
	}

	return sg.Result(steps), nil
}
