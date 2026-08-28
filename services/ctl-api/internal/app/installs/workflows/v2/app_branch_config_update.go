package v2

import (
	"encoding/json"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitinstallstackversionrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitrunnerhealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generateinstallstackversion"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generatestate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/updateappconfig"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

func AppBranchConfigUpdate(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	newAppConfigID := generics.FromPtrStr(flw.Metadata["new_app_config_id"])
	installConfigUpdateID := generics.FromPtrStr(flw.Metadata["install_config_update_id"])

	if newAppConfigID == "" {
		return nil, errors.New("new_app_config_id not found in workflow metadata")
	}

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	var diff *app.InstallConfigDiff
	if installConfigUpdateID != "" {
		diff, err = activities.AwaitGetInstallAppConfigVersionDiff(ctx, &activities.GetInstallAppConfigVersionDiffInput{
			InstallAppConfigVersionID: installConfigUpdateID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to get pre-computed config diff")
		}
	}

	appBranchRunID := generics.FromPtrStr(flw.Metadata["app_branch_run_id"])
	installGroupID := generics.FromPtrStr(flw.Metadata["install_group_id"])
	appReleaseID := generics.FromPtrStr(flw.Metadata["app_release_id"])
	var releaseComponentBuildIDs map[string]string
	releaseSandboxBuildID := ""
	if appReleaseID != "" {
		if err := json.Unmarshal([]byte(generics.FromPtrStr(flw.Metadata["release_component_build_ids"])), &releaseComponentBuildIDs); err != nil {
			return nil, errors.Wrap(err, "unable to decode release component builds")
		}
		releaseSandboxBuildID = generics.FromPtrStr(flw.Metadata["release_sandbox_build_id"])
	}
	releaseBuilds := WithReleaseBuilds(releaseComponentBuildIDs, releaseSandboxBuildID)
	triggeredBy := "api"
	if appBranchRunID != "" {
		triggeredBy = "app-branch"
	} else if appReleaseID != "" {
		triggeredBy = "app-release"
	}

	steps := make([]*app.WorkflowStep, 0)
	sg := newStepGroup(flw)

	sg.nextGroupEager()
	configStep, err := sg.installSignalStep(ctx, installID, "update app config", pgtype.Hstore{}, &updateappconfig.Signal{
		InstallID:                 installID,
		NewAppConfigID:            newAppConfigID,
		DryRun:                    flw.PlanOnly,
		AppBranchRunID:            appBranchRunID,
		InstallGroupID:            installGroupID,
		InstallAppConfigVersionID: installConfigUpdateID,
		TriggeredBy:               triggeredBy,
		Metadata:                  map[string]string{"source": triggeredBy},
	}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create update app config step")
	}
	steps = append(steps, configStep)

	sg.nextGroupEager()
	orgEnabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureStateGenV2))
	if err != nil {
		return nil, errors.Wrap(err, "unable to check state-gen-v2 feature")
	}
	stateGenV2 := statemanager.UseStateGenV2(orgEnabled, install.Metadata)

	if !stateGenV2 {
		step, err := sg.installSignalStep(ctx, installID, "generate install state", pgtype.Hstore{}, &generatestate.Signal{
			InstallID: installID,
		}, flw.PlanOnly, WithSkippable(false))
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	sg.nextGroupEager()
	step, err := sg.installSignalStep(ctx, installID, runnerHealthyStepName, pgtype.Hstore{}, &awaitrunnerhealthy.Signal{
		InstallID: installID,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	if diff != nil && diff.StackChanged {
		stackSteps, err := getStackVersionSteps(ctx, sg, installID, flw.PlanOnly)
		if err != nil {
			return nil, errors.Wrap(err, "unable to generate stack version steps")
		}
		steps = append(steps, stackSteps...)
	}

	if diff != nil && (diff.SandboxChanged || diff.SandboxBuildChanged) {
		flw.Metadata["skip_components"] = generics.ToPtr("true")

		newAppCfg, err := activities.AwaitGetAppConfigByID(ctx, newAppConfigID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get new app config")
		}

		awData, err := activities.AwaitGetActionWorkflows(ctx, &activities.GetActionWorkflows{
			InstallID: installID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to get action workflows")
		}

		dg := newGenCtx(sg, flw, installID, newAppCfg, awData, WithInstallInputs(install.CurrentInstallInputs), releaseBuilds)
		sandboxSteps, err := getSandboxReprovisionSteps(ctx, dg, install, sandboxNeedsRunnerHealthyGate(diff))
		if err != nil {
			return nil, errors.Wrap(err, "unable to generate sandbox reprovision steps")
		}
		steps = append(steps, sandboxSteps...)
	}

	newAppCfg, err := activities.AwaitGetAppConfigByID(ctx, newAppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get new app config")
	}

	awData, err := activities.AwaitGetActionWorkflows(ctx, &activities.GetActionWorkflows{
		InstallID: installID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get action workflows")
	}

	componentIDs, err := activities.AwaitGetAppGraph(ctx, activities.GetAppGraphRequest{
		InstallID: install.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install graph")
	}

	deployComponentIDs := filterComponentsByDiff(componentIDs, newAppCfg, diff)

	dg := newGenCtx(sg, flw, installID, newAppCfg, awData, WithInstallInputs(install.CurrentInstallInputs), releaseBuilds)
	deploySteps, err := getComponentDeploySteps(ctx, dg, deployComponentIDs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to generate component deploy steps")
	}
	steps = append(steps, deploySteps...)

	return sg.Result(steps), nil
}

// sandboxNeedsRunnerHealthyGate reports whether the sandbox reprovision phase has
// to wait on the runner again. This flow already waited before the stack steps,
// and only those steps can roll the runner out from under the sandbox — without
// them a second wait can only re-confirm the first, and rendered as a duplicate
// "runner healthy" step.
func sandboxNeedsRunnerHealthyGate(diff *app.InstallConfigDiff) bool {
	return diff != nil && diff.StackChanged
}

func filterComponentsByDiff(componentIDs []string, newAppCfg *app.AppConfig, diff *app.InstallConfigDiff) []string {
	newComponentSet := make(map[string]bool, len(newAppCfg.ComponentIDs))
	for _, id := range newAppCfg.ComponentIDs {
		newComponentSet[id] = true
	}

	if diff == nil {
		var filtered []string
		for _, id := range componentIDs {
			if newComponentSet[id] {
				filtered = append(filtered, id)
			}
		}
		return filtered
	}

	changedSet := make(map[string]bool)
	for _, e := range diff.Added {
		changedSet[e.ComponentID] = true
	}
	for _, e := range diff.Changed {
		changedSet[e.ComponentID] = true
	}

	var filtered []string
	for _, id := range componentIDs {
		if newComponentSet[id] && changedSet[id] {
			filtered = append(filtered, id)
		}
	}
	return filtered
}

// getStackVersionSteps emits a new install stack version and the wait for its run.
// This regenerates the stack template only — see getStackReprovisionSteps for the
// full stack recreation, which also recycles the runner service account and install
// state around it.
func getStackVersionSteps(ctx workflow.Context, sg *stepGroup, installID string, planOnly bool) ([]*app.WorkflowStep, error) {
	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, installID)
	if err != nil {
		return nil, err
	}

	var steps []*app.WorkflowStep

	sg.nextGroupEager()

	step, err := sg.installSignalStep(ctx, installID, "generate install stack", pgtype.Hstore{}, &generateinstallstackversion.Signal{
		InstallStackID: stack.ID,
	}, planOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = sg.installSignalStep(ctx, installID, "await install stack", pgtype.Hstore{}, &awaitinstallstackversionrun.Signal{
		InstallStackID: stack.ID,
	}, planOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	return steps, nil
}
