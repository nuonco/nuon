package v2

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitrunnerhealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generatestate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/installconfigdiff"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

// AppBranchConfigUpdate generates workflow steps for updating an install's app config
// as part of an app branch run. It diffs the install's current config against the new
// config and deploys only the changed components.
func AppBranchConfigUpdate(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	newAppConfigID := generics.FromPtrStr(flw.Metadata["new_app_config_id"])

	if newAppConfigID == "" {
		return nil, errors.New("new_app_config_id not found in workflow metadata")
	}

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	steps := make([]*app.WorkflowStep, 0)
	sg := newStepGroup(flw)

	// Step 1: Compute config diff between current and new app config
	sg.nextGroupEager()
	step, err := sg.installSignalStep(ctx, installID, "config diff", pgtype.Hstore{}, &installconfigdiff.Signal{
		InstallID:      installID,
		NewAppConfigID: newAppConfigID,
	}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create config diff step")
	}
	steps = append(steps, step)

	// Generate install state
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

	// Wait for runner healthy
	sg.nextGroupEager()
	step, err = sg.installSignalStep(ctx, installID, "runner healthy", pgtype.Hstore{}, &awaitrunnerhealthy.Signal{
		InstallID: installID,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	// Load new app config to get the target component set
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

	// Get component IDs in dependency order
	componentIDs, err := activities.AwaitGetAppGraph(ctx, activities.GetAppGraphRequest{
		InstallID: install.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install graph")
	}

	// Filter to only components present in the new config
	newComponentSet := make(map[string]bool, len(newAppCfg.ComponentIDs))
	for _, id := range newAppCfg.ComponentIDs {
		newComponentSet[id] = true
	}

	var deployComponentIDs []string
	for _, id := range componentIDs {
		if newComponentSet[id] {
			deployComponentIDs = append(deployComponentIDs, id)
		}
	}

	// Deploy changed components
	dg := newGenCtx(sg, flw, installID, newAppCfg, awData)
	deploySteps, err := getComponentDeploySteps(ctx, dg, deployComponentIDs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to generate component deploy steps")
	}
	steps = append(steps, deploySteps...)

	return sg.Result(steps), nil
}
