package v2

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitinstallstackversionrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitrunnerhealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generateinstallstackversion"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/updateappconfig"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

const appBranchGroupDirectiveVersion = "app-branch-group-directive-v1"

func AppBranchConfigUpdate(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	newAppConfigID := generics.FromPtrStr(flw.Metadata["new_app_config_id"])
	installConfigUpdateID := generics.FromPtrStr(flw.Metadata["install_config_update_id"])

	if newAppConfigID == "" {
		return nil, errors.New("new_app_config_id not found in workflow metadata")
	}

	useGroupDirective := workflow.GetVersion(ctx, appBranchGroupDirectiveVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion
	if useGroupDirective {
		if err := classifyAndMaybeWait(ctx, flw, installID, newAppConfigID, installConfigUpdateID); err != nil {
			return nil, err
		}
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

	steps := make([]*app.WorkflowStep, 0)
	sg := newStepGroup(flw)

	sg.nextGroupEager()
	configStep, err := sg.installSignalStep(ctx, installID, "update app config", pgtype.Hstore{}, &updateappconfig.Signal{
		InstallID:      installID,
		NewAppConfigID: newAppConfigID,
		DryRun:         flw.PlanOnly,
		AppBranchRunID: appBranchRunID,
		InstallGroupID: installGroupID,
		TriggeredBy:    "app-branch",
		Metadata:       map[string]string{"source": "app-branch"},
	}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create update app config step")
	}
	steps = append(steps, configStep)

	if useGroupDirective && installConfigDiffEmpty(diff) {
		return sg.Result(steps), nil
	}

	stackChanged := diff != nil && diff.StackChanged

	// A stack change recycles the runner, so gating on the outgoing one would
	// block the apply that brings its replacement up.
	if !stackChanged {
		sg.nextGroupEager()
		step, err := sg.installSignalStep(ctx, installID, runnerHealthyStepName, pgtype.Hstore{}, &awaitrunnerhealthy.Signal{
			InstallID: installID,
			Mode:      awaitrunnerhealthy.ModeRequireActive,
		}, flw.PlanOnly)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	if stackChanged {
		stackSteps, err := getStackVersionSteps(ctx, sg, installID, flw.PlanOnly)
		if err != nil {
			return nil, errors.Wrap(err, "unable to generate stack version steps")
		}
		steps = append(steps, stackSteps...)

		sg.nextGroup()
		step, err := sg.installSignalStep(ctx, installID, runnerHealthyStepName, pgtype.Hstore{}, &awaitrunnerhealthy.Signal{
			InstallID: installID,
			Mode:      awaitrunnerhealthy.ModeStartup,
		}, flw.PlanOnly)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create post-stack runner health step")
		}
		steps = append(steps, step)
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

		dg := newGenCtx(sg, flw, installID, newAppCfg, awData, WithInstallInputs(install.CurrentInstallInputs))
		sandboxSteps, err := getSandboxReprovisionSteps(ctx, dg, install, false)
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

	// Order against the config being rolled out, not the one the install is
	// still pinned to: a component this update adds has no vertex in the old
	// graph, and would otherwise never get a deploy step.
	componentIDs, err := activities.AwaitGetAppGraph(ctx, activities.GetAppGraphRequest{
		InstallID:   install.ID,
		AppConfigID: newAppConfigID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install graph")
	}

	deployComponentIDs := filterComponentsByDiff(componentIDs, newAppCfg, diff)

	dg := newGenCtx(sg, flw, installID, newAppCfg, awData, WithInstallInputs(install.CurrentInstallInputs))
	deploySteps, err := getComponentDeploySteps(ctx, dg, deployComponentIDs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to generate component deploy steps")
	}
	steps = append(steps, deploySteps...)

	return sg.Result(steps), nil
}

func classifyAndMaybeWait(ctx workflow.Context, flw *app.Workflow, installID, newAppConfigID, installConfigUpdateID string) error {
	decision, err := activities.AwaitClassifyInstallGroupDirective(ctx, activities.ClassifyInstallGroupDirectiveRequest{
		InstallID:      installID,
		WorkflowID:     flw.ID,
		NewAppConfigID: newAppConfigID,
		AppBranchRunID: generics.FromPtrStr(flw.Metadata["app_branch_run_id"]),
	})
	if err != nil {
		return errors.Wrap(err, "unable to classify install group directive")
	}
	if decision.Directive == "release" {
		if _, err := activities.AwaitSendInstallGroupDirective(ctx, activities.SendInstallGroupDirectiveRequest{
			GroupWorkflowID: generics.FromPtrStr(flw.Metadata["group_workflow_id"]),
			Namespace:       generics.FromPtrStr(flw.Metadata["group_workflow_namespace"]),
			InstallID:       installID,
			AppBranchRunID:  generics.FromPtrStr(flw.Metadata["app_branch_run_id"]),
			Directive:       decision.Directive,
			Reason:          decision.Reason,
			WaitingOnRunID:  decision.WaitingOnRunID,
		}); err != nil {
			return errors.Wrap(err, "unable to send install group directive")
		}
	}
	if !decision.WaitForPrior {
		return nil
	}
	for {
		terminal, err := activities.AwaitInstallUpdateTerminal(ctx, activities.InstallUpdateTerminalRequest{
			WorkflowID: decision.PriorWorkflowID,
		})
		if err != nil {
			return errors.Wrap(err, "unable to check prior install update")
		}
		if terminal.Terminal {
			break
		}
		if err := workflow.Sleep(ctx, 15*time.Second); err != nil {
			return err
		}
	}
	if _, err := activities.AwaitRecomputeInstallConfigDiff(ctx, activities.RecomputeInstallConfigDiffRequest{
		InstallID:                 installID,
		InstallAppConfigVersionID: installConfigUpdateID,
		NewAppConfigID:            newAppConfigID,
	}); err != nil {
		return errors.Wrap(err, "unable to recompute install config diff")
	}
	return nil
}

func installConfigDiffEmpty(diff *app.InstallConfigDiff) bool {
	if diff == nil {
		return false
	}
	return !diff.StackChanged && !diff.SandboxChanged && !diff.SandboxBuildChanged &&
		len(diff.Added) == 0 && len(diff.Changed) == 0 && len(diff.Removed) == 0
}

// filterComponentsByDiff narrows a dependency-ordered component list to the ones
// this config update touches. componentIDs must be ordered against newAppCfg —
// the result can only ever be a subset of it, so anything the update adds is
// silently dropped if the caller ordered against the install's current config.
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

	step, err = sg.installSignalStep(ctx, installID, app.AwaitInstallStackStepName, pgtype.Hstore{}, &awaitinstallstackversionrun.Signal{
		InstallStackID:     stack.ID,
		CreateManagedStack: true,
	}, planOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	return steps, nil
}
