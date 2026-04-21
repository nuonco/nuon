package workflows

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
	appconfig "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/appconfig"
	builds "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/builds"
	deploygrouptoqueue "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/deploygrouptoqueue"
	fetchcommit "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/fetchcommit"
	sandboxbuild "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/sandboxbuild"
)

// AppBranchRun builds the workflow steps for an app branch run
// This workflow orchestrates:
// 1. Fetching the latest commit from VCS
// 2. Cloning the repo and parsing the intermediate config
// 3. Building all components in the config
// 4. Deploying to install groups in order
func AppBranchRun(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	// Extract metadata from workflow
	appBranchID := generics.FromPtrStr(flw.Metadata["app_branch_id"])
	if appBranchID == "" {
		return nil, errors.New("app_branch_id not found in workflow metadata")
	}

	runID := generics.FromPtrStr(flw.Metadata["run_id"])
	if runID == "" {
		return nil, errors.New("run_id not found in workflow metadata")
	}

	configID := generics.FromPtrStr(flw.Metadata["config_id"])
	if configID == "" {
		return nil, errors.New("config_id not found in workflow metadata")
	}

	steps := make([]*app.WorkflowStep, 0)
	sg := newStepGroup()

	// Step 1: Fetch commit from VCS and store on run
	sg.nextGroup()
	step, err := sg.appBranchSignalStep(ctx, appBranchID, "fetch commit", pgtype.Hstore{}, &fetchcommit.Signal{
		RunID:       runID,
		AppBranchID: appBranchID,
	}, WithSkippable(false))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create fetch commit step")
	}
	steps = append(steps, step)

	// Step 2: Clone the repo and parse intermediate config
	sg.nextGroup()
	step, err = sg.appBranchSignalStep(ctx, appBranchID, "fetch app config", pgtype.Hstore{}, &appconfig.Signal{
		AppBranchID: appBranchID,
		RunID:       runID,
	}, WithSkippable(false))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create app config step")
	}
	steps = append(steps, step)

	// Step 3: Build all components and sandbox builds in parallel
	sg.nextGroup()
	step, err = sg.appBranchSignalStep(ctx, appBranchID, "builds", pgtype.Hstore{}, &builds.Signal{
		AppBranchID: appBranchID,
		RunID:       runID,
	}, WithSkippable(false))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create builds step")
	}
	steps = append(steps, step)

	// Step 3.5: Build sandbox (conditional — only if an AppSandboxConfig exists for this app)
	step, err = sg.appBranchSignalStep(ctx, appBranchID, "build sandbox", pgtype.Hstore{}, &sandboxbuild.Signal{
		AppBranchID: appBranchID,
		RunID:       runID,
	}, WithSkippable(false))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create sandbox build step")
	}
	steps = append(steps, step)

	// Step 4: Deploy to install groups in order
	// Fetch install groups for this config, ordered by the order field
	installGroups, err := activities.AwaitGetInstallGroupsByConfigID(ctx, configID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch install groups")
	}

	// Create sequential steps for each install group
	for _, group := range installGroups {
		sg.nextGroup()
		step, err = sg.appBranchSignalStep(ctx, appBranchID, "deploy install group: "+group.Name, pgtype.Hstore{}, &deploygrouptoqueue.Signal{
			InstallGroupID: group.ID,
			AppBranchID:    appBranchID,
			RunID:          runID,
		}, WithSkippable(false))
		if err != nil {
			return nil, errors.Wrapf(err, "unable to create deploy step for group %s", group.Name)
		}
		steps = append(steps, step)
	}

	return sg.Result(steps), nil
}
