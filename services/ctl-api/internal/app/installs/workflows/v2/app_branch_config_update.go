package v2

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitinstallstackversionrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitrunnerhealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generateinstallstackversion"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generatestate"
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
		diff, err = activities.AwaitGetInstallConfigUpdateDiff(ctx, &activities.GetInstallConfigUpdateDiffInput{
			InstallConfigUpdateID: installConfigUpdateID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to get pre-computed config diff")
		}
	}

	steps := make([]*app.WorkflowStep, 0)
	sg := newStepGroup(flw)

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
	step, err := sg.installSignalStep(ctx, installID, "runner healthy", pgtype.Hstore{}, &awaitrunnerhealthy.Signal{
		InstallID: installID,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	if diff != nil && diff.StackChanged {
		stackSteps, err := getStackReprovisionSteps(ctx, sg, installID, flw.PlanOnly)
		if err != nil {
			return nil, errors.Wrap(err, "unable to generate stack reprovision steps")
		}
		steps = append(steps, stackSteps...)
	}

	if diff != nil && diff.SandboxChanged {
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
		sandboxSteps, err := getSandboxReprovisionSteps(ctx, dg, install)
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

	dg := newGenCtx(sg, flw, installID, newAppCfg, awData, WithInstallInputs(install.CurrentInstallInputs))
	deploySteps, err := getComponentDeploySteps(ctx, dg, deployComponentIDs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to generate component deploy steps")
	}
	steps = append(steps, deploySteps...)

	return sg.Result(steps), nil
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

func getStackReprovisionSteps(ctx workflow.Context, sg *stepGroup, installID string, planOnly bool) ([]*app.WorkflowStep, error) {
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
