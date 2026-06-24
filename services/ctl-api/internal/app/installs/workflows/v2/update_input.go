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

	sg.nextGroup() // refresh inputs partial in state
	step, err := stateInputsRefreshStep(ctx, sg, install, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

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

	// Get all components that reference the changed inputs
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

	// Reconcile toggleable-component enable/disable transitions carried by the
	// synthetic enabled inputs. Enabling a currently-inactive component fires
	// its enable lifecycle around the deploy; disabling a currently-active one
	// tears it down with its disable lifecycle; disabling something that is not
	// deployed is a no-op skip.
	enableComps, disableComps, skipComps, err := classifyEnabledTransitions(ctx, dg, appConfig, changedInputs)
	if err != nil {
		return nil, err
	}
	// Components transitioning to effectively-enabled must be deployed even if
	// they were not otherwise pulled in by the changed-input/dependent scan
	// (e.g. a dependent re-enabled only because its dependency came back).
	componentIDs = append(componentIDs, enableComps...)
	componentIDs = generics.UniqueSlice(componentIDs)
	// Components being torn down or skipped must not be deployed.
	componentIDs = removeComponentIDs(componentIDs, disableComps, skipComps)
	// Deploy dependencies before the components that depend on them.
	componentIDs = dg.topoSort(componentIDs)

	// Check if sandbox config references contain any of the changed inputs
	sandboxNeedsReprovision, err := checkSandboxNeedsReprovision(ctx, appConfig, &changedRefs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check if sandbox needs reprovision")
	}

	preEnableSteps, err := componentEnableLifecycleSteps(ctx, dg, enableComps, app.ActionWorkflowTriggerTypePreEnableComponent)
	if err != nil {
		return nil, err
	}
	steps = append(steps, preEnableSteps...)

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

// classifyEnabledTransitions inspects the changed synthetic enabled inputs and
// reconciles the effective-enabled state of every affected component into the
// transitions to act on. Toggling a component changes not only its own
// effective state but that of everything that depends on it (declared deps or
// output refs), so the affected set is each directly-toggled component plus its
// transitive dependents. For each affected component we compare its desired
// effective-enabled state against whether it is currently deployed:
//   - enable: effectively enabled but not currently deployed (deploy + enable hooks)
//   - disable: effectively disabled but currently deployed (teardown + disable hooks)
//   - skip: a directly-toggled component that is disabled and not deployed (no-op)
//
// A component that is effectively enabled and already deployed is left
// untouched here; it flows through the regular deploy path as a routine
// redeploy with no enable lifecycle, preserving "enable hooks fire only on an
// off→on transition". Cascaded dependents that end up disabled-and-undeployed
// are dropped silently rather than emitting noisy per-dependent skip steps.
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

	affected := dg.transitiveDependentsClosure(toggledIDs)
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

// componentEnableLifecycleSteps emits the per-component enable lifecycle action
// steps for the given components and trigger. Enable hooks never run on a
// plan-only workflow.
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

// componentDisableSteps emits teardown steps (wrapped in disable lifecycle
// hooks) for each disabled-and-deployed component, plus a discoverable no-op
// skip step for each disabled-but-undeployed component.
func componentDisableSteps(ctx workflow.Context, dg *genCtx, disableComps, skipComps []string) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0)

	for _, compID := range disableComps {
		comp, ok := dg.components[compID]
		if !ok {
			continue
		}

		// Disable hooks never run on a plan-only workflow, mirroring the enable
		// side; the teardown itself is still emitted plan-only so the plan shows
		// the destroy.
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
			"reason": generics.ToPtr("component is already not deployed on this install"),
		}, nil, false)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create disable skip step")
		}
		steps = append(steps, skipStep)
	}

	return steps, nil
}

// removeComponentIDs returns ids with every member of the removal sets dropped,
// preserving order.
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

	// Per-component install-level overrides (Helm values / Terraform vars) are
	// carried as reserved synthetic inputs that target a single component by name
	// rather than via config refs. Decode the targeted component names so editing
	// an override redeploys exactly that component (no ref injection required).
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

// checkSandboxNeedsReprovision checks if the sandbox configuration references any of the changed inputs
func checkSandboxNeedsReprovision(ctx workflow.Context, appCfg *app.AppConfig, changedRefs *[]refs.Ref) (bool, error) {
	// Check if any of the sandbox's references match the changed inputs
	for _, sandboxRef := range appCfg.SandboxConfig.Refs {
		for _, changedRef := range *changedRefs {
			if sandboxRef.Name == changedRef.Name && sandboxRef.Type == changedRef.Type {
				return true, nil
			}
		}
	}

	return false, nil
}
