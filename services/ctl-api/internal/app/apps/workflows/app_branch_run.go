package workflows

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
	appconfig "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/appconfig"
	buildcomponents "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/buildcomponents"
	checkchanges "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/checkchanges"
	deploygrouptoqueue "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/deploygrouptoqueue"
)

// AppBranchRun builds the workflow steps for an app branch run
// This workflow orchestrates:
// 1. Fetching the repo and building the config
// 2. Building all components in the config
// 3. Deploying to install groups in order
func AppBranchRun(ctx workflow.Context, flw *app.Workflow) ([]*app.WorkflowStep, error) {
	// Extract metadata from workflow
	appBranchID := generics.FromPtrStr(flw.Metadata["app_branch_id"])
	if appBranchID == "" {
		return nil, errors.New("app_branch_id not found in workflow metadata")
	}

	configID := generics.FromPtrStr(flw.Metadata["config_id"])
	if configID == "" {
		return nil, errors.New("config_id not found in workflow metadata")
	}

	steps := make([]*app.WorkflowStep, 0)
	sg := newStepGroup()

	// Step 1: Check for changes and update app config
	sg.nextGroup()
	step, err := sg.appBranchSignalStep(ctx, appBranchID, "check for changes", pgtype.Hstore{}, &checkchanges.Signal{
		AppBranchID: appBranchID,
	}, WithSkippable(false))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create check changes step")
	}
	steps = append(steps, step)

	sg.nextGroup()
	step, err = sg.appBranchSignalStep(ctx, appBranchID, "fetch app config", pgtype.Hstore{}, &appconfig.Signal{
		AppBranchID: appBranchID,
		CommitSHA:   "", // TODO: Get commit SHA from workflow metadata or branch
	}, WithSkippable(false))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create app config step")
	}
	steps = append(steps, step)

	// Step 2: Build all components in parallel
	sg.nextGroup()
	step, err = sg.appBranchSignalStep(ctx, appBranchID, "build all components", pgtype.Hstore{}, &buildcomponents.Signal{
		AppBranchID: appBranchID,
	}, WithSkippable(false))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create build components step")
	}
	steps = append(steps, step)

	// Step 3: Deploy to install groups in order
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
		}, WithSkippable(false))
		if err != nil {
			return nil, errors.Wrapf(err, "unable to create deploy step for group %s", group.Name)
		}
		steps = append(steps, step)
	}

	return steps, nil
}
