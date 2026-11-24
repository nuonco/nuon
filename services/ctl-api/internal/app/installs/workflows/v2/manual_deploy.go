package v2

import (
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitcomponenthealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitrunnerhealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentdeployapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentdeploysyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentsyncimage"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func ManualDeploySteps(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	sg := newStepGroup(flw)

	steps := make([]*app.WorkflowStep, 0)

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}
	if err := failIfPinnedAppConfigMoved(install, flw); err != nil {
		return nil, err
	}
	targets, err := flw.ExecutionTargets()
	if err != nil {
		return nil, err
	}

	sg.nextGroupEager()
	step, err := sg.installSignalStep(ctx, installID, "runner healthy", pgtype.Hstore{}, &awaitrunnerhealthy.Signal{
		InstallID: installID,
		Mode:      awaitrunnerhealthy.ModeRequireActive,
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
	primaryBuildID := ""
	if targets != nil {
		primaryBuildID = targets.PrimaryBuildID
		if installDeploy.OrgID != targets.OrgID || installDeploy.InstallComponentID != targets.InstallComponents[targets.PrimaryComponentID] || installDeploy.ComponentBuildID != targets.PrimaryBuildID || generics.FromPtrStr(installDeploy.InstallWorkflowID) != flw.ID || installDeploy.AppConfigID != targets.AppConfigID {
			return nil, errors.New("primary deploy does not match prepared manual deploy target")
		}
		if err := hydrateRunbookComponentTargets(appCfg, targets); err != nil {
			return nil, err
		}
		selected := make([]*app.InstallActionWorkflow, 0, len(targets.Actions))
		for _, action := range awData {
			if target, ok := targets.Actions[action.ActionWorkflowID]; ok && target.InstallActionWorkflowID == action.ID {
				if action.InstallID != targets.InstallID || action.OrgID != targets.OrgID || action.ActionWorkflow.AppID != targets.AppID || action.ActionWorkflow.OrgID != targets.OrgID {
					return nil, errors.New("lifecycle action does not belong to prepared manual deploy")
				}
				selected = append(selected, action)
			}
		}
		if len(selected) != len(targets.Actions) {
			return nil, errors.New("prepared lifecycle action is no longer installed")
		}
		awData = selected
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
	if targets != nil {
		if comp.ID != targets.PrimaryComponentID || comp.AppID != targets.AppID || comp.OrgID != targets.OrgID {
			return nil, errors.New("component does not belong to prepared manual deploy")
		}
		if err := targets.CheckComponent(comp.ID, installComp); err != nil {
			return nil, err
		}
		comp.Type = targets.PrimaryType
		if target, ok := targets.Components[comp.ID]; ok {
			comp.Name = target.Name
		}
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
	if unsupportedComponentType(*comp) {
		sg.nextGroup()
		skipStep, err := unsupportedComponentTypeSkipStep(ctx, sg, installID, *comp, "skipped deploy "+comp.Name)
		if err != nil {
			return nil, err
		}
		steps = append(steps, skipStep)
		return sg.Result(steps), nil
	}

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
		deployStep, err := sg.installSignalStep(ctx, installID, "sync "+comp.Name, componentStepMetadata(comp.Name), &componentsyncimage.Signal{
			InstallComponentID: installComp.ID,
			InstallID:          installID,
			DeployID:           generics.FromPtrStr(installDeployID),
			BuildID:            primaryBuildID,
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
		planStep, err := sg.installSignalStep(ctx, installID, "sync and plan "+comp.Name, componentStepMetadata(comp.Name), &componentdeploysyncandplan.Signal{
			InstallComponentID: installComp.ID,
			InstallID:          installID,
			DeployID:           generics.FromPtrStr(installDeployID),
			BuildID:            primaryBuildID,
			ComponentID:        comp.ID,
			FlowID:             "",
			SandboxMode:        false,
			Role:               flw.Role,
		}, flw.PlanOnly)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create image sync")
		}
		applyPlanStep, err := sg.installSignalStep(ctx, installID, "apply "+comp.Name, componentStepMetadata(comp.Name), &componentdeployapplyplan.Signal{
			InstallComponentID: installComp.ID,
			InstallID:          installID,
			ComponentID:        comp.ID,
			BuildID:            primaryBuildID,
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
			if componentGateEnabled(ctx, installID, comp.ID, comp.Type) {
				gateStep, err := sg.installSignalStep(ctx, installID, "verify health "+comp.Name, componentStepMetadata(comp.Name), &awaitcomponenthealthy.Signal{
					InstallID:          installID,
					InstallComponentID: installComp.ID,
					ComponentID:        comp.ID,
				}, flw.PlanOnly, WithSkippable(false), WithMaxAutoRetries(3))
				if err != nil {
					return nil, errors.Wrap(err, "unable to create verify health step")
				}
				steps = append(steps, gateStep)
			}
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
	var dependencyCompIDs []string
	if targets != nil {
		if generics.FromPtrStr(deployDependents) != strconv.FormatBool(true) {
			return sg.Result(steps), nil
		}
		dependencyCompIDs = targets.ComponentOrder
	} else {
		componentIDs, err := activities.AwaitGetAppComponentGraph(ctx, activities.GetAppComponentGraphRequest{
			InstallID: install.ID, ComponentID: comp.ID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to get app component graph")
		}
		dependencyCompIDs = generics.SliceAfterValue(componentIDs, comp.ID)
	}
	dependencyDeploySteps, err := getComponentDeploySteps(ctx, dg, dependencyCompIDs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component deploy steps")
	}

	if generics.FromPtrStr(deployDependents) == strconv.FormatBool(true) {
		steps = append(steps, dependencyDeploySteps...)
	}

	return sg.Result(steps), nil
}
