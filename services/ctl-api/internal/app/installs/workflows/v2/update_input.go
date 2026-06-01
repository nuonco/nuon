package v2

import (
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generatestate"
	statepartialgenerate "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/state/statepartialgenerate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
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
	orgEnabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureStateGenV2))
	if err != nil {
		return nil, errors.Wrap(err, "unable to check state-gen-v2 feature")
	}
	stateGenV2 := statemanager.UseStateGenV2(orgEnabled, install.Metadata)
	var stateSignal signal.Signal
	if stateGenV2 {
		stateSignal = &statepartialgenerate.Signal{
			InstallID:       installID,
			Targets:         statemanager.TargetsForHint(statemanager.HintInputsUpdated, ""),
			TriggeredByID:   installID,
			TriggeredByType: "installs",
		}
	} else {
		stateSignal = &generatestate.Signal{InstallID: installID}
	}
	step, err := sg.installSignalStep(ctx, installID, "update install state inputs", pgtype.Hstore{}, stateSignal, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	changedInputsRaw := generics.FromPtrStr(flw.Metadata["inputs"])
	changedInputs := strings.Split(changedInputsRaw, ",")

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

	lifecycleSteps, err := getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreUpdateInputs, sg, appConfig, awData)
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

		dependentCompIDs, err := activities.AwaitGetComponentDependents(ctx, &activities.GetComponentDependentsRequest{
			AppConfigID: appConfig.ID,
			ComponentID: comp.ID,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "unable to get component dependents for %s", comp.ID)
		}

		componentIDs = append(componentIDs, dependentCompIDs...)
	}
	componentIDs = generics.UniqueSlice(componentIDs)

	// Check if sandbox config references contain any of the changed inputs
	sandboxNeedsReprovision, err := checkSandboxNeedsReprovision(ctx, appConfig, &changedRefs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check if sandbox needs reprovision")
	}

	// If sandbox needs reprovision, add sandbox reprovision steps before component deploys
	if sandboxNeedsReprovision {
		sandboxSteps, err := getSandboxReprovisionSteps(ctx, install, installID, flw, sg, appConfig, awData)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get sandbox reprovision steps")
		}
		steps = append(steps, sandboxSteps...)
	} else {
		deploySteps, err := getComponentDeploySteps(ctx, installID, flw, componentIDs, sg, appConfig, awData)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get component deploy steps")
		}
		steps = append(steps, deploySteps...)
	}

	lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostUpdateInputs, sg, appConfig, awData)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	return sg.Result(steps), nil
}

func getComponentsForChangedInputs(appConfig *app.AppConfig, changedRefs *[]refs.Ref) []app.Component {
	components := make([]app.Component, 0)

	for _, conConfigs := range appConfig.ComponentConfigConnections {
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
