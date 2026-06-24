package workflows

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	appconfig "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/appconfig"
	builds "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/builds"
	fetchcommit "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/fetchcommit"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/planinstallgroup"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/updateinstallgroup"
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

	appConfigID := generics.FromPtrStr(flw.Metadata["app_config_id"])
	skipBuilds := generics.FromPtrStr(flw.Metadata["skip_builds"]) == "true"

	steps := make([]*app.WorkflowStep, 0)
	sg := newStepGroup()

	if appConfigID == "" {
		// Normal flow: fetch commit and parse config from VCS
		sg.nextGroup()
		step, err := sg.appBranchSignalStep(ctx, appBranchID, "fetch commit", pgtype.Hstore{}, &fetchcommit.Signal{
			RunID:       runID,
			AppBranchID: appBranchID,
		}, WithSkippable(false))
		if err != nil {
			return nil, errors.Wrap(err, "unable to create fetch commit step")
		}
		steps = append(steps, step)

		sg.nextGroup()
		step, err = sg.appBranchSignalStep(ctx, appBranchID, "fetch app config", pgtype.Hstore{}, &appconfig.Signal{
			AppBranchID: appBranchID,
			RunID:       runID,
		}, WithSkippable(false))
		if err != nil {
			return nil, errors.Wrap(err, "unable to create app config step")
		}
		steps = append(steps, step)
	} else {
		// Pre-existing app config: skip VCS fetch and config parse
		sg.nextGroup()
		step, err := sg.appBranchSignalStep(ctx, appBranchID, "fetch commit (skipped)", pgtype.Hstore{}, nil)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create skipped fetch commit step")
		}
		steps = append(steps, step)

		sg.nextGroup()
		step, err = sg.appBranchSignalStep(ctx, appBranchID, "fetch app config (skipped)", pgtype.Hstore{}, nil)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create skipped app config step")
		}
		steps = append(steps, step)
	}

	// Step 3: Build all components and sandbox
	if appConfigID != "" && skipBuilds {
		sg.nextGroup()
		step, err := sg.appBranchSignalStep(ctx, appBranchID, "building components and sandbox (skipped)", pgtype.Hstore{}, nil)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create skipped builds step")
		}
		steps = append(steps, step)
	} else {
		sg.nextGroup()
		step, err := sg.appBranchSignalStep(ctx, appBranchID, "building components and sandbox", pgtype.Hstore{}, &builds.Signal{
			AppBranchID: appBranchID,
			RunID:       runID,
		}, WithSkippable(false))
		if err != nil {
			return nil, errors.Wrap(err, "unable to create builds step")
		}
		steps = append(steps, step)
	}

	// Step 4: Deploy to install groups in order
	// Fetch install groups for this config, ordered by the order field
	allInstallGroups, err := activities.AwaitGetInstallGroupsByConfigID(ctx, configID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch install groups")
	}

	// For plan-only (preview) runs, only include groups marked UseForPreviews.
	// If none are marked, fall back to the first group.
	installGroups := allInstallGroups
	if flw.PlanOnly && len(allInstallGroups) > 0 {
		var previewGroups []*app.AppBranchInstallGroup
		for _, g := range allInstallGroups {
			if g.UseForPreviews {
				previewGroups = append(previewGroups, g)
			}
		}
		if len(previewGroups) > 0 {
			installGroups = previewGroups
		} else {
			installGroups = allInstallGroups[:1]
		}
	}

	// Create plan + deploy steps for each install group
	for _, group := range installGroups {
		sg.nextGroup()
		planStep, err := sg.appBranchSignalStep(ctx, appBranchID, "plan install group: "+group.Name, pgtype.Hstore{}, &planinstallgroup.Signal{
			InstallGroupID: group.ID,
			AppBranchID:    appBranchID,
			RunID:          runID,
		}, WithExecutionType(app.WorkflowStepExecutionTypeApproval))
		if err != nil {
			return nil, errors.Wrapf(err, "unable to create plan step for group %s", group.Name)
		}
		steps = append(steps, planStep)

		sg.nextGroup()
		deployStep, err := sg.appBranchSignalStep(ctx, appBranchID, "deploy install group: "+group.Name, pgtype.Hstore{}, &updateinstallgroup.Signal{
			InstallGroupID: group.ID,
			AppBranchID:    appBranchID,
			RunID:          runID,
		}, WithExecutionType(app.WorkflowStepExecutionTypeApproval), WithSkippable(true))
		if err != nil {
			return nil, errors.Wrapf(err, "unable to create deploy step for group %s", group.Name)
		}
		steps = append(steps, deployStep)
	}

	return sg.Result(steps), nil
}
