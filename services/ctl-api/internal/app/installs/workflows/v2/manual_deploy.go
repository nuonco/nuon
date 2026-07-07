package v2

import (
	"strconv"

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

func ManualDeploySteps(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	sg := newStepGroup(flw)

	steps := make([]*app.WorkflowStep, 0)

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	sg.nextGroupEager() // generate install state
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

	sg.nextGroupEager()
	step, err := sg.installSignalStep(ctx, installID, "runner healthy", pgtype.Hstore{}, &awaitrunnerhealthy.Signal{
		InstallID: installID,
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
	deployDependencies := flw.Metadata["deploy_dependencies"]

	installDeploy, err := activities.AwaitGetDeployByDeployID(ctx, generics.FromPtrStr(installDeployID))
	if err != nil {
		return nil, errors.New("unable to get install deploy")
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

	dg := newGenCtx(sg, flw, installID, appCfg, awData, WithInstallInputs(install.CurrentInstallInputs))

	// first, provision the deploy with before and after triggers
	comp, err := activities.AwaitGetComponentByComponentID(ctx, installDeploy.ComponentID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component")
	}

	// Resolve install component ID for v2 signals
	installComp, err := activities.AwaitGetInstallComponent(ctx, activities.GetInstallComponentRequest{
		InstallID:   installID,
		ComponentID: comp.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install component")
	}

	// When the primary is a non-image component, walk its image deps and
	// prepend sync steps for any whose latest Active build differs from
	// what's currently deployed. The deploy_components path does this
	// inside getComponentDeploySteps; the manual_deploy path emits the
	// primary inline, so we have to drive the same logic here.
	//
	// This MUST run before the pre-deploy actions below: those actions can
	// read the dependency's image outputs, so the new image bytes have to
	// land in the install registry first (matches deploy_components
	// ordering). When dep syncs are emitted they occupy their own step
	// group, and we start a fresh group so pre-deploy actions run strictly
	// after them. When nothing needs syncing we leave the group untouched
	// so pre-deploy actions stay in the existing group (no empty group).
	//
	// dg.addedImageDepSyncs is shared with the dependents call below so a
	// shared image dep is only synced once across both call sites.
	if !flw.PlanOnly && !comp.Type.IsImage() && generics.FromPtrStr(deployDependencies) == strconv.FormatBool(true) {
		depSyncSteps, err := getImageDepSyncSteps(ctx, dg, comp.ID, 0, map[string]int{comp.ID: 0})
		if err != nil {
			return nil, errors.Wrap(err, "unable to prepend image-dep sync steps for primary")
		}
		if len(depSyncSteps) > 0 {
			steps = append(steps, depSyncSteps...)
			sg.nextGroup()
		}
	}

	preDeploySteps, err := getComponentLifecycleActionsSteps(ctx, dg, comp, app.ActionWorkflowTriggerTypePreDeployComponent)
	if err != nil {
		return nil, err
	}
	if !flw.PlanOnly {
		steps = append(steps, preDeploySteps...)
	}

	// sync image
	if comp.Type.IsImage() {
		sg.nextGroup() // component sync
		deployStep, err := sg.installSignalStep(ctx, installID, "sync "+comp.Name, pgtype.Hstore{}, &componentsyncimage.Signal{
			InstallComponentID: installComp.ID,
			DeployID:           generics.FromPtrStr(installDeployID),
			ComponentID:        comp.ID,
			FlowID:             "",
			SandboxMode:        false,
			Role:               flw.Role,
		}, flw.PlanOnly)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create image sync")
		}

		steps = append(steps, deployStep)

		// Record that this image component's sync is already emitted so the
		// subsequent dependents pass (getComponentDeploySteps below) won't
		// emit a duplicate "sync X (dep)" step when a dependent lists this
		// image as a dep.
		dg.addedImageDepSyncs[comp.ID] = struct{}{}
	} else {
		sg.nextGroup() // component sync + plan + apply
		planStep, err := sg.installSignalStep(ctx, installID, "sync and plan "+comp.Name, pgtype.Hstore{}, &componentdeploysyncandplan.Signal{
			InstallComponentID: installComp.ID,
			InstallID:          installID,
			DeployID:           generics.FromPtrStr(installDeployID),
			ComponentID:        comp.ID,
			FlowID:             "",
			SandboxMode:        false,
			Role:               flw.Role,
		}, flw.PlanOnly, WithSkippable(false))
		if err != nil {
			return nil, errors.Wrap(err, "unable to create image sync")
		}
		applyPlanStep, err := sg.installSignalStep(ctx, installID, "apply "+comp.Name, pgtype.Hstore{}, &componentdeployapplyplan.Signal{
			InstallComponentID: installComp.ID,
			InstallID:          installID,
			ComponentID:        comp.ID,
			FlowID:             "",
			SandboxMode:        false,
		}, flw.PlanOnly, WithMaxAutoRetries(componentMaxAutoRetries(appCfg, comp.ID)))
		if err != nil {
			return nil, errors.Wrap(err, "unable to create image sync")
		}

		if flw.PlanOnly {
			steps = append(steps, planStep)
		} else {
			steps = append(steps, planStep, applyPlanStep)
		}
	}

	postDeploySteps, err := getComponentLifecycleActionsSteps(ctx, dg, comp, app.ActionWorkflowTriggerTypePostDeployComponent)
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
	dependencyDeploySteps, err := getComponentDeploySteps(ctx, dg, dependencyCompIDs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component deploy steps")
	}

	if generics.FromPtrStr(deployDependents) == strconv.FormatBool(true) {
		steps = append(steps, dependencyDeploySteps...)
	}

	return sg.Result(steps), nil
}
