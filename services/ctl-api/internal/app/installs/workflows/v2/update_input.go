package v2

import (
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitrunnerhealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func InputUpdate(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	sg := newStepGroup(flw)
	steps := make([]*app.WorkflowStep, 0)

	sg.nextGroupEager()
	step, err := sg.installSignalStep(ctx, installID, runnerHealthyStepName, pgtype.Hstore{}, &awaitrunnerhealthy.Signal{
		InstallID: installID,
		Mode:      awaitrunnerhealthy.ModeRequireActive,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	sg.nextGroup()
	step, err = stateInputsRefreshStep(ctx, sg, install, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	if flw.IsInputsOnly() {
		return sg.Result(steps), nil
	}

	changedInputsRaw := generics.FromPtrStr(flw.Metadata["inputs"])
	changedInputs := strings.Split(changedInputsRaw, ",")
	deployDependents := generics.FromPtrStr(flw.Metadata["deploy_dependents"]) == strconv.FormatBool(true)

	appConfig, err := activities.AwaitGetAppConfig(ctx, activities.GetAppConfigRequest{
		ID: install.AppConfigID,
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

	dg := newGenCtx(sg, flw, installID, appConfig, awData, WithInstallInputs(install.CurrentInstallInputs))

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

	var componentIDs []string
	for _, comp := range getComponentsForChangedInputs(appConfig, &changedRefs, changedInputs) {
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

	enableComps, disableComps, skipComps, err := classifyEnabledTransitions(ctx, dg, appConfig, changedInputs)
	if err != nil {
		return nil, err
	}
	// why: Components transitioning to effectively-enabled must be deployed even if
	// they were not otherwise pulled in by the changed-input/dependent scan
	// (e.g. a dependent re-enabled only because its dependency came back).
	componentIDs = append(componentIDs, enableComps...)
	componentIDs = generics.UniqueSlice(componentIDs)
	componentIDs = removeComponentIDs(componentIDs, disableComps, skipComps)
	componentIDs = dg.topoSort(componentIDs)

	sandboxNeedsReprovision, err := checkSandboxNeedsReprovision(ctx, appConfig, &changedRefs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check if sandbox needs reprovision")
	}

	preEnableSteps, err := componentEnableLifecycleSteps(ctx, dg, enableComps, app.ActionWorkflowTriggerTypePreEnableComponent)
	if err != nil {
		return nil, err
	}
	steps = append(steps, preEnableSteps...)

	if sandboxNeedsReprovision {
		sandboxSteps, err := getSandboxReprovisionSteps(ctx, dg, install, false)
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

	postEnableSteps, err := componentEnableLifecycleSteps(ctx, dg, enableComps, app.ActionWorkflowTriggerTypePostEnableComponent)
	if err != nil {
		return nil, err
	}
	steps = append(steps, postEnableSteps...)

	disableSteps, err := componentDisableSteps(ctx, dg, disableComps, skipComps)
	if err != nil {
		return nil, err
	}
	steps = append(steps, disableSteps...)

	lifecycleSteps, err = getLifecycleActionsSteps(ctx, dg, app.ActionWorkflowTriggerTypePostUpdateInputs)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	return sg.Result(steps), nil
}

// why: classifyEnabledTransitions inspects the changed synthetic enabled inputs and
// reconciles the effective-enabled state of every directly-toggled component
// into the transitions to act on. We act only on the components whose own
// synthetic enabled input changed; dependents are deliberately NOT cascaded.
// An inconsistent desired state (an enabled component depending on a disabled
// one, or a disabled component with enabled dependents) is rejected up front by
// the installvalidate framework, so by the time this runs only the toggled
// component itself needs to transition. For each toggled component we compare
// its desired effective-enabled state against whether it is currently deployed:
//   - enable: effectively enabled but not currently deployed (deploy + enable hooks)
//   - disable: effectively disabled but currently deployed (teardown + disable hooks)
//   - skip: a directly-toggled component that is disabled and not deployed (no-op)
//
// A component that is effectively enabled and already deployed is left
// untouched here; it flows through the regular deploy path as a routine
// redeploy with no enable lifecycle, preserving "enable hooks fire only on an
// off→on transition".
//
// enable is returned in deploy order (dependencies first); disable in teardown
// order (dependents first).
func classifyEnabledTransitions(ctx workflow.Context, dg *genCtx, appConfig *app.AppConfig, changedInputs []string) (enable, disable, skip []string, err error) {
	cccByName := make(map[string]*app.ComponentConfigConnection, len(appConfig.ComponentConfigConnections))
	for i := range appConfig.ComponentConfigConnections {
		ccc := &appConfig.ComponentConfigConnections[i]
		cccByName[ccc.Component.Name] = ccc
	}

	toggledSet := make(map[string]struct{})
	var toggledIDs []string
	for _, input := range changedInputs {
		kind, compName, ok := config.ParseComponentOverrideInputName(input)
		if !ok || kind != config.ComponentOverrideKindEnabled {
			continue
		}
		ccc, ok := cccByName[compName]
		if !ok || !ccc.IsToggleable() {
			continue
		}
		if _, dup := toggledSet[ccc.ComponentID]; dup {
			continue
		}
		toggledSet[ccc.ComponentID] = struct{}{}
		toggledIDs = append(toggledIDs, ccc.ComponentID)
	}
	if len(toggledIDs) == 0 {
		return nil, nil, nil, nil
	}

	affected := toggledIDs
	installComps, err := activities.AwaitGetInstallComponentsBatch(ctx, activities.GetInstallComponentsBatchRequest{
		InstallID:    dg.installID,
		ComponentIDs: affected,
	})
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "unable to batch get install components for toggle reconcile")
	}

	for _, compID := range affected {
		effEnabled := dg.effectiveEnabled(compID)
		active := false
		if ic, ok := installComps[compID]; ok && ic != nil {
			active = ic.Status == app.InstallComponentStatusActive
		}
		switch {
		case effEnabled && !active:
			enable = append(enable, compID)
		case !effEnabled && active:
			disable = append(disable, compID)
		case !effEnabled && !active:
			if _, ok := toggledSet[compID]; ok {
				skip = append(skip, compID)
			}
		}
	}

	enable = dg.topoSort(enable)
	disable = dg.reverseTopoSort(disable)
	return enable, disable, skip, nil
}

func componentEnableLifecycleSteps(ctx workflow.Context, dg *genCtx, compIDs []string, trigger app.ActionWorkflowTriggerType) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0)
	if dg.flw.PlanOnly {
		return steps, nil
	}
	for _, compID := range compIDs {
		comp, ok := dg.components[compID]
		if !ok {
			continue
		}
		s, err := getComponentLifecycleActionsSteps(ctx, dg, &comp, trigger)
		if err != nil {
			return nil, err
		}
		steps = append(steps, s...)
	}
	return steps, nil
}

func componentDisableSteps(ctx workflow.Context, dg *genCtx, disableComps, skipComps []string) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0)

	for _, compID := range disableComps {
		comp, ok := dg.components[compID]
		if !ok {
			continue
		}

		if !dg.flw.PlanOnly {
			preDisable, err := getComponentLifecycleActionsSteps(ctx, dg, &comp, app.ActionWorkflowTriggerTypePreDisableComponent)
			if err != nil {
				return nil, err
			}
			steps = append(steps, preDisable...)
		}

		teardownSteps, err := getComponentTeardownSteps(ctx, dg, comp)
		if err != nil {
			return nil, err
		}
		steps = append(steps, teardownSteps...)

		if !dg.flw.PlanOnly {
			postDisable, err := getComponentLifecycleActionsSteps(ctx, dg, &comp, app.ActionWorkflowTriggerTypePostDisableComponent)
			if err != nil {
				return nil, err
			}
			steps = append(steps, postDisable...)
		}
	}

	for _, compID := range skipComps {
		comp, ok := dg.components[compID]
		if !ok {
			continue
		}
		dg.sg.nextGroup()
		skipStep, err := dg.sg.installSignalStep(ctx, dg.installID, "skipped disable "+comp.Name, pgtype.Hstore{
			"reason":         generics.ToPtr("component is already not deployed on this install"),
			"component_name": generics.ToPtr(comp.Name),
		}, nil, false)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create disable skip step")
		}
		steps = append(steps, skipStep)
	}

	return steps, nil
}

func removeComponentIDs(ids []string, removeSets ...[]string) []string {
	remove := make(map[string]struct{})
	for _, set := range removeSets {
		for _, id := range set {
			remove[id] = struct{}{}
		}
	}
	if len(remove) == 0 {
		return ids
	}
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, drop := remove[id]; !drop {
			out = append(out, id)
		}
	}
	return out
}

func getComponentsForChangedInputs(appConfig *app.AppConfig, changedRefs *[]refs.Ref, changedInputs []string) []app.Component {
	components := make([]app.Component, 0)

	overrideTargets := make(map[string]struct{})
	for _, name := range changedInputs {
		if _, comp, ok := config.ParseComponentOverrideInputName(name); ok {
			overrideTargets[comp] = struct{}{}
		}
	}

	for _, conConfigs := range appConfig.ComponentConfigConnections {
		if _, ok := overrideTargets[conConfigs.Component.Name]; ok {
			components = append(components, conConfigs.Component)
			continue
		}
		for _, ref := range conConfigs.Refs {
			for _, changedRef := range *changedRefs {
				if ref.Name == changedRef.Name && ref.Type == changedRef.Type {
					components = append(components, conConfigs.Component)
				}
			}
		}
	}
	return components
}

func checkSandboxNeedsReprovision(ctx workflow.Context, appCfg *app.AppConfig, changedRefs *[]refs.Ref) (bool, error) {
	for _, sandboxRef := range appCfg.SandboxConfig.Refs {
		for _, changedRef := range *changedRefs {
			if sandboxRef.Name == changedRef.Name && sandboxRef.Type == changedRef.Type {
				return true, nil
			}
		}
	}

	return false, nil
}
