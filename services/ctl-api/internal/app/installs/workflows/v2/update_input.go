package v2

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func InputUpdate(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	changedInputsRaw := generics.FromPtrStr(flw.Metadata["inputs"])
	changedInputs := strings.Split(changedInputsRaw, ",")
	deployDependents := generics.FromPtrStr(flw.Metadata["deploy_dependents"]) == strconv.FormatBool(true)

	sg := newStepGroup(flw)
	steps, err := inputUpdateSteps(ctx, sg, inputUpdateStepsArgs{
		Install:          install,
		Flw:              flw,
		ChangedInputs:    changedInputs,
		DeployDependents: deployDependents,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get app config for install %s", installID)
	}

	awData, err := activities.AwaitGetActionWorkflows(ctx, &activities.GetActionWorkflows{
		InstallID: installID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get action workflows")
	}

	dg := newGenCtx(sg, flw, installID, appConfig, awData)

	lifecycleSteps, err := getLifecycleActionsSteps(ctx, dg, app.ActionWorkflowTriggerTypePreUpdateInputs)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	var changedRefs []refs.Ref
	for _, input := range changedInputs {
		changedRefs = append(changedRefs, refs.Ref{
			Name: input,
			Type: refs.RefTypeInputs,
		})
		changedRefs = append(changedRefs, refs.Ref{
			Name: input,
			Type: refs.RefTypeInstallInputs,
		})
	}

	// Get all components that reference the changed inputs
	var componentIDs []string
	for _, comp := range getComponentsForChangedInputs(appConfig, &changedRefs) {
		componentIDs = append(componentIDs, comp.ID)

		if deployDependents {
			dependentCompIDs, err := activities.AwaitGetComponentDependents(ctx, &activities.GetComponentDependentsRequest{
				AppConfigID: appConfig.ID,
				ComponentID: comp.ID,
			})
			if err != nil {
				return nil, errors.Wrapf(err, "unable to get component dependents for %s", comp.ID)
			}

			componentIDs = append(componentIDs, dependentCompIDs...)
		}
	}
	componentIDs = generics.UniqueSlice(componentIDs)

	// Check if sandbox config references contain any of the changed inputs
	sandboxNeedsReprovision, err := checkSandboxNeedsReprovision(ctx, appConfig, &changedRefs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check if sandbox needs reprovision")
	}

	// If sandbox needs reprovision, add sandbox reprovision steps before component deploys
	if sandboxNeedsReprovision {
		sandboxSteps, err := getSandboxReprovisionSteps(ctx, dg, install)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get sandbox reprovision steps")
		}
		steps = append(steps, sandboxSteps...)
	} else {
		deploySteps, err := getComponentDeploySteps(ctx, dg, componentIDs)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get component deploy steps")
		}
		steps = append(steps, deploySteps...)
	}

	lifecycleSteps, err = getLifecycleActionsSteps(ctx, dg, app.ActionWorkflowTriggerTypePostUpdateInputs)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	return sg.Result(steps), nil
}
