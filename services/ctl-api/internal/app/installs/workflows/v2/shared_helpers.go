package v2

import (
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/actionworkflowrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitrunnerhealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentdeployapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentdeploysyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentsyncimage"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentteardownapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentteardownsyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/executeactionworkflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

// genCtx is the per-workflow-invocation context for step generation. It bundles
// the read-only inputs every helper needs (workflow row, install ID, pinned
// app config, install action workflows) along with the workflow-scoped
// mutable state (step group, image-dep sync dedup map, derived component
// maps).
//
// One genCtx is constructed at workflow entry and threaded through every
// helper that emits steps. Sharing it across all calls inside a single
// workflow gives dep-aware deploys a single dedup boundary for free.
type genCtx struct {
	sg        *stepGroup
	flw       *app.Workflow
	installID string
	appCfg    *app.AppConfig
	awData    []*app.InstallActionWorkflow

	// enabledInputs holds the install's latest input values, used to resolve
	// whether a toggleable component is enabled via its synthetic enabled input
	// (see config.EnabledOverrideInputName). May be nil if no inputs exist.
	enabledInputs map[string]*string

	// Derived once from appCfg.ComponentConfigConnections for dep-aware
	// deploys.
	components   map[string]app.Component
	depIDsByComp map[string][]string
	cccByComp    map[string]*app.ComponentConfigConnection

	// enablement resolves effective-enabled state and cascade ordering from the
	// pinned app config and the install's enabled inputs.
	enablement *app.ComponentEnablementResolver

	// addedImageDepSyncs tracks image-dep components that already had a
	// sync step prepended in this workflow. Multiple non-image components
	// may share the same image dep; we only sync it once per workflow.
	addedImageDepSyncs map[string]struct{}
}

func newGenCtx(sg *stepGroup, flw *app.Workflow, installID string, appCfg *app.AppConfig, awData []*app.InstallActionWorkflow, opts ...genCtxOption) *genCtx {
	components, depIDsByComp, cccByComp := buildComponentConfigMaps(appCfg)
	dg := &genCtx{
		sg:                 sg,
		flw:                flw,
		installID:          installID,
		appCfg:             appCfg,
		awData:             awData,
		components:         components,
		depIDsByComp:       depIDsByComp,
		cccByComp:          cccByComp,
		addedImageDepSyncs: make(map[string]struct{}),
	}
	for _, opt := range opts {
		opt(dg)
	}
	dg.enablement = app.NewComponentEnablementResolver(dg.cccByComp, dg.enabledInputs)
	return dg
}

type genCtxOption func(*genCtx)

// WithInstallInputs supplies the install's latest input values so the step
// generator can resolve toggleable-component enabled-state from the synthetic
// enabled inputs (the source of truth for component toggles).
func WithInstallInputs(ii *app.InstallInputs) genCtxOption {
	return func(dg *genCtx) {
		if ii != nil {
			dg.enabledInputs = ii.Values
		}
	}
}

// componentEnabledFromInputs resolves whether a toggleable component is enabled
// from a set of install input values. The synthetic enabled input
// (config.EnabledOverrideInputName) is the source of truth; when unset it falls
// back to the component's default_enabled. Non-toggleable components are always
// enabled.
func componentEnabledFromInputs(enabledInputs map[string]*string, ccc *app.ComponentConfigConnection) bool {
	return app.ComponentEnabledFromInputs(enabledInputs, ccc)
}

// buildComponentConfigMaps builds the (components, depIDsByComp, cccByComp)
// trio used by the dep-aware deploy logic from an AppConfig's pinned
// ComponentConfigConnections. cccByComp records membership of each component
// in the install's pinned app config snapshot — used to skip deps that are
// not part of this app config version. The build-to-push for each dep is
// resolved at workflow generation time via the GetLatestActiveComponentBuild
// activity (global heuristic — no cross-ACV pinning).
func buildComponentConfigMaps(appCfg *app.AppConfig) (
	map[string]app.Component,
	map[string][]string,
	map[string]*app.ComponentConfigConnection,
) {
	components := make(map[string]app.Component, len(appCfg.ComponentConfigConnections))
	depIDsByComp := make(map[string][]string, len(appCfg.ComponentConfigConnections))
	cccByComp := make(map[string]*app.ComponentConfigConnection, len(appCfg.ComponentConfigConnections))
	for i := range appCfg.ComponentConfigConnections {
		ccc := &appCfg.ComponentConfigConnections[i]
		components[ccc.ComponentID] = ccc.Component
		depIDsByComp[ccc.ComponentID] = []string(ccc.ComponentDependencyIDs)
		cccByComp[ccc.ComponentID] = ccc
	}
	return components, depIDsByComp, cccByComp
}

// filterActionWorkflowsByTrigger filters pre-fetched install action workflows by trigger type,
// optionally scoped to a specific component. It uses the version-pinned configs from
// appCfg.ActionWorkflowConfigs (with Triggers preloaded) instead of fetching latest configs.
// This replaces individual AwaitGetInstallActionWorkflowsByTriggerType activity calls
// with in-memory filtering.
func filterActionWorkflowsByTrigger(installActionWorkflows []*app.InstallActionWorkflow, triggerTyp app.ActionWorkflowTriggerType, componentID string, appCfg *app.AppConfig) []*app.InstallActionWorkflow {
	awcMap := make(map[string]app.ActionWorkflowConfig, len(appCfg.ActionWorkflowConfigs))
	for _, awc := range appCfg.ActionWorkflowConfigs {
		awcMap[awc.ActionWorkflowID] = awc
	}

	indices := map[string]int{}
	wkflows := make(map[string]*app.InstallActionWorkflow, len(installActionWorkflows))

	for _, wf := range installActionWorkflows {
		cfg, ok := awcMap[wf.ActionWorkflowID]
		if !ok {
			continue
		}

		if componentID == "" {
			if cfg.HasTrigger(triggerTyp) {
				wkflows[wf.ID] = wf
				indices[wf.ID] = cfg.GetTriggerIndex(triggerTyp)
			}
		} else {
			if cfg.HasComponentTrigger(triggerTyp, componentID) {
				wkflows[wf.ID] = wf
				indices[wf.ID] = cfg.GetComponentTriggerIndex(triggerTyp, componentID)
			}
		}
	}

	workflowIDs := make([]string, 0, len(indices))
	for wkflowID := range indices {
		workflowIDs = append(workflowIDs, wkflowID)
	}

	sort.SliceStable(workflowIDs, func(i, j int) bool {
		return indices[workflowIDs[i]] < indices[workflowIDs[j]]
	})

	result := make([]*app.InstallActionWorkflow, 0, len(workflowIDs))
	for _, wkflowID := range workflowIDs {
		if wf, ok := wkflows[wkflowID]; ok {
			result = append(result, wf)
		}
	}

	return result
}

func getComponentLifecycleActionsSteps(ctx workflow.Context, dg *genCtx, comp *app.Component, triggerTyp app.ActionWorkflowTriggerType) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0)
	installActions := filterActionWorkflowsByTrigger(dg.awData, triggerTyp, comp.ID, dg.appCfg)

	for _, installAction := range installActions {
		sig := &executeactionworkflow.Signal{
			Signal: &actionworkflowrun.Signal{
				InstallID:               dg.installID,
				InstallActionWorkflowID: installAction.ID,
				TriggerType:             triggerTyp,
				TriggeredByID:           dg.flw.ID,
				TriggeredByType:         string(triggerTyp),
				RunEnvVars: map[string]string{
					"TRIGGER_TYPE":   string(triggerTyp),
					"COMPONENT_ID":   comp.ID,
					"COMPONENT_NAME": comp.Name,
				},
				Role: dg.flw.Role,
			},
		}
		name := fmt.Sprintf("%s Action Run (%s)", installAction.ActionWorkflow.Name, triggerTyp)
		step, err := dg.sg.installSignalStep(ctx, dg.installID, name, pgtype.Hstore{}, sig, dg.flw.PlanOnly)
		if err != nil {
			return nil, err
		}

		steps = append(steps, step)
	}

	return steps, nil
}

func getLifecycleActionsSteps(ctx workflow.Context, dg *genCtx, triggerTyp app.ActionWorkflowTriggerType) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0)
	installActions := filterActionWorkflowsByTrigger(dg.awData, triggerTyp, "", dg.appCfg)

	if len(installActions) == 0 {
		return steps, nil
	}

	dg.sg.nextGroup() // lifecycleSteps

	for _, installAction := range installActions {
		sig := &executeactionworkflow.Signal{
			Signal: &actionworkflowrun.Signal{
				InstallID:               dg.installID,
				InstallActionWorkflowID: installAction.ID,
				TriggerType:             triggerTyp,
				TriggeredByID:           dg.flw.ID,
				TriggeredByType:         string(triggerTyp),
				RunEnvVars: map[string]string{
					"TRIGGER_TYPE": string(triggerTyp),
					"FLOW_TYPE":    string(dg.flw.Type),
					"FLOW_ID":      dg.flw.ID,
					// TODO(sdboyer) remove these once they're updated on the other end
					"INSTALL_WORKFLOW_TYPE": string(dg.flw.Type),
					"INSTALL_WORKFLOW_ID":   dg.flw.ID,
				},
				Role: dg.flw.Role,
			},
		}
		name := fmt.Sprintf("%s Action Run (%s)", installAction.ActionWorkflow.Name, triggerTyp)
		step, err := dg.sg.installSignalStep(ctx, dg.installID, name, pgtype.Hstore{}, sig, dg.flw.PlanOnly)
		if err != nil {
			return nil, err
		}

		steps = append(steps, step)
	}

	return steps, nil
}

func getComponentDeploySteps(ctx workflow.Context, dg *genCtx, componentIDs []string) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0)

	// componentIDIdx maps each component ID to its position in componentIDs.
	// Used by the dep-aware deploy logic to avoid prepending an image sync
	// step for an image dep that already appears earlier in this batch (it
	// will be sync'd by the normal per-component flow).
	componentIDIdx := make(map[string]int, len(componentIDs))
	for i, id := range componentIDs {
		componentIDIdx[id] = i
	}

	// Batch fetch all install components in one activity call instead of
	// fetching them individually per component in the loop below.
	installComps, err := activities.AwaitGetInstallComponentsBatch(ctx, activities.GetInstallComponentsBatchRequest{
		InstallID:    dg.installID,
		ComponentIDs: componentIDs,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to batch get install components")
	}

	for i, compID := range componentIDs {
		// Yield to the Temporal scheduler periodically to avoid deadlock detection
		// when generating steps for many components (TMPRL1101).
		// Sleep(0) is a no-op; a real timer is needed to force a yield.
		if i%5 == 0 && i > 0 {
			_ = workflow.Sleep(ctx, time.Millisecond)
		}

		comp, has := dg.components[compID]
		if !has {
			return nil, errors.Errorf("component %s not found in app config", compID)
		}

		// Skip a component that is not effectively enabled: either its own
		// toggle is off, or a component it depends on (declared or
		// output-referenced) is disabled. This guards every deploy path — not
		// just toggle reconciliation — from deploying a component against a
		// dependency that is gone.
		if _, ok := dg.cccByComp[compID]; ok && !dg.effectiveEnabled(compID) {
			continue
		}

		// Dep-aware image-sync prepend.
		//
		// When deploying a non-image component, walk its image dependencies
		// and prepend a sync step for any image dep whose latest Active
		// ComponentBuild differs from what is currently deployed on the
		// install. The dep sync steps live in their own step group ordered
		// before the parent component's group so the new image bytes land in
		// the install registry before the parent renders/deploys against
		// them.
		//
		// Skipped when:
		//   - dg.flw.PlanOnly is true (matches existing image-sync gating)
		//   - the parent component is itself an image (its own sync is
		//     handled by the normal flow below)
		if !dg.flw.PlanOnly && !comp.Type.IsImage() {
			depSyncSteps, err := getImageDepSyncSteps(ctx, dg, compID, i, componentIDIdx)
			if err != nil {
				return nil, err
			}
			steps = append(steps, depSyncSteps...)
		}

		dg.sg.nextGroup()

		var installComponentID string
		if installComp, ok := installComps[compID]; ok && installComp != nil {
			installComponentID = installComp.ID
		}

		if !dg.flw.PlanOnly {
			preDeploySteps, err := getComponentLifecycleActionsSteps(ctx, dg, &comp, app.ActionWorkflowTriggerTypePreDeployComponent)
			if err != nil {
				return nil, err
			}
			steps = append(steps, preDeploySteps...)
		}

		// sync image
		if comp.Type.IsImage() && !dg.flw.PlanOnly {
			// Resolve the build to push using the global "latest active
			// build for this component" heuristic. This does NOT enforce
			// per-AppConfig-version pinning — a deploy of an install
			// pinned to ACV vN may pick up a build created against vN+1
			// or vN-1. Cross-ACV correctness is a deferred fix.
			//
			// When no Active build exists yet (e.g. the build is still
			// in-flight), leave BuildID empty: the signal falls back to
			// resolving the latest build at run time. This keeps a single
			// component without an active build from aborting generation of
			// the entire deploy workflow.
			latestBuild, err := activities.AwaitGetLatestActiveComponentBuildByComponentID(ctx, comp.ID)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to resolve latest active build for image component %s", comp.Name)
			}
			var buildID string
			if latestBuild != nil {
				buildID = latestBuild.ID
			}
			deployStep, err := dg.sg.installSignalStep(ctx, dg.installID, "sync "+comp.Name, pgtype.Hstore{}, &componentsyncimage.Signal{
				InstallComponentID: installComponentID,
				ComponentID:        comp.ID,
				BuildID:            buildID,
				Role:               dg.flw.Role,
			}, dg.flw.PlanOnly)
			if err != nil {
				return nil, errors.Wrap(err, "unable to create image sync")
			}

			steps = append(steps, deployStep)
		} else {
			if dg.flw.PlanOnly && comp.Type == app.ComponentTypeExternalImage || comp.Type == app.ComponentTypeDockerBuild {
				continue
			}

			// Resolve the build to deploy using the global heuristic.
			// Same caveat as the image-sync branch above — cross-ACV
			// build leakage is possible. When no Active build exists yet,
			// leave BuildID empty so the signal resolves the latest build
			// at run time rather than aborting the whole workflow.
			latestBuild, err := activities.AwaitGetLatestActiveComponentBuildByComponentID(ctx, comp.ID)
			if err != nil {
				return nil, errors.Wrapf(err, "unable to resolve latest active build for component %s", comp.Name)
			}
			var buildID string
			if latestBuild != nil {
				buildID = latestBuild.ID
			}
			planStep, err := dg.sg.installSignalStep(ctx, dg.installID, "sync and plan "+comp.Name, pgtype.Hstore{}, &componentdeploysyncandplan.Signal{
				InstallComponentID: installComponentID,
				InstallID:          dg.installID,
				ComponentID:        comp.ID,
				BuildID:            buildID,
				Role:               dg.flw.Role,
			}, dg.flw.PlanOnly, WithSkippable(false))
			if err != nil {
				return nil, errors.Wrap(err, "unable to create image sync")
			}

			applyPlanStep, err := dg.sg.installSignalStep(ctx, dg.installID, "apply "+comp.Name, pgtype.Hstore{}, &componentdeployapplyplan.Signal{
				InstallComponentID: installComponentID,
				InstallID:          dg.installID,
				ComponentID:        comp.ID,
			}, dg.flw.PlanOnly, WithMaxAutoRetries(componentMaxAutoRetries(dg.appCfg, comp.ID)))
			if err != nil {
				return nil, errors.Wrap(err, "unable to create image sync")
			}
			if dg.flw.PlanOnly {
				steps = append(steps, planStep)
			} else {
				steps = append(steps, planStep, applyPlanStep)
			}
		}
		if !dg.flw.PlanOnly {
			postDeploySteps, err := getComponentLifecycleActionsSteps(ctx, dg, &comp, app.ActionWorkflowTriggerTypePostDeployComponent)
			if err != nil {
				return nil, err
			}
			steps = append(steps, postDeploySteps...)
		}
	}

	return steps, nil
}

// getComponentTeardownSteps emits the steps that tear a single component down
// off an install: a teardown sync-and-plan followed by an apply for
// infrastructure components, or a no-op skip step for image components (whose
// bytes live in the registry and have nothing to destroy). It opens its own
// step group and does not include lifecycle action steps — callers weave those
// in around it.
func getComponentTeardownSteps(ctx workflow.Context, dg *genCtx, comp app.Component) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0)
	dg.sg.nextGroup()

	if comp.Type.IsImage() {
		skipStep, err := dg.sg.installSignalStep(ctx, dg.installID, "skipped image disable "+comp.Name, pgtype.Hstore{
			"reason": generics.ToPtr("skipped image teardown"),
		}, nil, false)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create image teardown skip step")
		}
		return append(steps, skipStep), nil
	}

	installComp, err := activities.AwaitGetInstallComponent(ctx, activities.GetInstallComponentRequest{
		InstallID:   dg.installID,
		ComponentID: comp.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install component")
	}

	planStep, err := dg.sg.installSignalStep(ctx, dg.installID, "teardown sync and plan "+comp.Name, pgtype.Hstore{}, &componentteardownsyncandplan.Signal{
		InstallComponentID: installComp.ID,
		InstallID:          dg.installID,
		ComponentID:        comp.ID,
		Role:               dg.flw.Role,
	}, dg.flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create teardown sync and plan step")
	}

	applyStep, err := dg.sg.installSignalStep(ctx, dg.installID, "teardown apply plan "+comp.Name, pgtype.Hstore{}, &componentteardownapplyplan.Signal{
		InstallComponentID: installComp.ID,
		InstallID:          dg.installID,
		ComponentID:        comp.ID,
	}, dg.flw.PlanOnly, WithMaxAutoRetries(componentMaxAutoRetries(dg.appCfg, comp.ID)))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create teardown apply step")
	}

	if dg.flw.PlanOnly {
		return append(steps, planStep), nil
	}
	return append(steps, planStep, applyStep), nil
}

// getImageDepSyncSteps returns image-dep sync steps to prepend before the
// non-image parent component identified by parentCompID at parentIdx in
// componentIDs. It walks the parent's pinned dependencies (from the AppConfig
// snapshot), filters to image-typed deps, and emits a componentsyncimage
// signal for each image dep whose latest Active ComponentBuild — for the
// install's pinned app config version — differs from the build currently
// deployed on the install.
//
// Why the lookup is pinned to the install's app config version:
//
// Each ComponentConfigConnection (ccc) is the per-app-config-version snapshot
// of a component's config and is what owns that component's builds. When an
// app config is re-synced, fresh ccc rows are created for every component;
// builds that happen later are tied to the new ccc, not the old. The dep-sync
// decision must therefore use the ccc that belongs to the consumer's pinned
// app config version. Asking for "latest active build for component X" across
// every ccc (i.e. across every app config version) over-syncs into installs
// pinned to an older app config and conceptually decouples the dep from the
// consumer's snapshot.
//
// The returned steps are added to a single step group ordered before the
// parent's group: the function calls sg.nextGroup() lazily on the first
// emitted step so that no empty group is left behind when no dep needs
// syncing.
//
// Dedup rules:
//   - skip image deps already handled earlier in this workflow (dg.addedImageDepSyncs)
//   - skip image deps that already appear earlier in componentIDs (their
//     normal per-component sync step will run earlier in the workflow)
//
// Quiet skip rules (no error returned, no step emitted):
//   - dep is not present in the loaded app config
//   - dep is not an image-typed component
//   - dep has no ccc in the install's pinned app config snapshot (the app
//     config doesn't include this dep at all)
//   - install component for the dep does not exist (nothing to sync against)
//   - dep has no Active ComponentBuild yet for the pinned ccc
//   - dep's currently deployed ComponentBuildID matches the latest Active
//     build for the pinned ccc (nothing to do)
func getImageDepSyncSteps(
	ctx workflow.Context,
	dg *genCtx,
	parentCompID string,
	parentIdx int,
	componentIDIdx map[string]int,
) ([]*app.WorkflowStep, error) {
	depIDs := dg.depIDsByComp[parentCompID]
	if len(depIDs) == 0 {
		return nil, nil
	}

	steps := make([]*app.WorkflowStep, 0)
	groupStarted := false

	for _, depID := range depIDs {
		if _, already := dg.addedImageDepSyncs[depID]; already {
			continue
		}

		// Skip if the dep is being deployed earlier in this same batch:
		// its normal per-component image-sync step will run before the
		// parent's group anyway, and adding another would double-sync.
		if depIdx, in := componentIDIdx[depID]; in && depIdx < parentIdx {
			continue
		}

		dep, has := dg.components[depID]
		if !has {
			// Dep is not part of this app config snapshot — nothing to do.
			continue
		}
		if !dep.Type.IsImage() {
			continue
		}

		// Membership check only: skip deps that aren't part of this app
		// config snapshot. We don't use the CCC for build resolution —
		// see note below on cross-ACV leakage.
		if depCCC, hasCCC := dg.cccByComp[depID]; !hasCCC || depCCC == nil {
			continue
		}

		// Resolve the dep's latest Active build via the global heuristic
		// (most recent Active build for this component across all CCCs).
		// This can leak a build from a newer/older AppConfig version into
		// the sync decision — cross-ACV correctness is a deferred fix.
		// Skip silently when no Active build exists yet — the next deploy
		// attempt will pick it up once the build lands.
		latestActive, err := activities.AwaitGetLatestActiveComponentBuildByComponentID(ctx, depID)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to resolve latest active build for image dep %s", depID)
		}
		if latestActive == nil {
			continue
		}

		depInstallComp, err := activities.AwaitGetInstallComponent(ctx, activities.GetInstallComponentRequest{
			InstallID:   dg.installID,
			ComponentID: depID,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "unable to get install component for image dep %s", depID)
		}
		if depInstallComp == nil {
			// No install component record yet for this dep — there is no
			// runner-side state to sync against. The normal install
			// bootstrapping flow is responsible for creating it; skip
			// silently here.
			continue
		}

		// AwaitGetInstallComponent preloads the most recent InstallDeploy
		// (any type, ORDER BY created_at DESC LIMIT 1). For image
		// components every install_deploy is a sync-image, so the most
		// recent deploy is the currently-synced build. When the deployed
		// build matches the app-config-version-pinned latest Active
		// build, no sync is needed.
		var deployedBuildID string
		if len(depInstallComp.InstallDeploys) > 0 {
			deployedBuildID = depInstallComp.InstallDeploys[0].ComponentBuildID
		}
		if deployedBuildID == latestActive.ID {
			continue
		}

		if !groupStarted {
			dg.sg.nextGroup()
			groupStarted = true
		}

		step, err := dg.sg.installSignalStep(ctx, dg.installID, "sync "+dep.Name+" (dep)", pgtype.Hstore{}, &componentsyncimage.Signal{
			InstallComponentID: depInstallComp.ID,
			ComponentID:        dep.ID,
			BuildID:            latestActive.ID,
			Role:               dg.flw.Role,
		}, dg.flw.PlanOnly)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to create dep image sync for %s", depID)
		}

		steps = append(steps, step)
		dg.addedImageDepSyncs[depID] = struct{}{}
	}

	return steps, nil
}

func deployAllComponents(ctx workflow.Context, dg *genCtx) ([]*app.WorkflowStep, error) {
	componentIDs, err := activities.AwaitGetAppGraph(ctx, activities.GetAppGraphRequest{
		InstallID: dg.installID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install graph")
	}

	steps := make([]*app.WorkflowStep, 0)

	step, err := dg.sg.installSignalStep(ctx, dg.installID, "runner healthy", pgtype.Hstore{}, &awaitrunnerhealthy.Signal{
		InstallID: dg.installID,
	}, dg.flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	var lifecycleSteps []*app.WorkflowStep
	if !dg.flw.PlanOnly {
		lifecycleSteps, err = getLifecycleActionsSteps(ctx, dg, app.ActionWorkflowTriggerTypePreDeployAllComponents)
		if err != nil {
			return nil, err
		}
		steps = append(steps, lifecycleSteps...)
	}
	deploySteps, err := getComponentDeploySteps(ctx, dg, componentIDs)
	if err != nil {
		return nil, err
	}
	steps = append(steps, deploySteps...)
	if !dg.flw.PlanOnly {
		// Yield after processing all component deploy steps to avoid deadlock detection (TMPRL1101).
		_ = workflow.Sleep(ctx, time.Millisecond)

		lifecycleSteps, err = getLifecycleActionsSteps(ctx, dg, app.ActionWorkflowTriggerTypePostDeployAllComponents)
		if err != nil {
			return nil, err
		}
		steps = append(steps, lifecycleSteps...)
	}

	return steps, nil
}
