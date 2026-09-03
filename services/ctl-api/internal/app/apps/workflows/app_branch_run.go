package workflows

import (
	"encoding/json"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	appconfig "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/appconfig"
	builds "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/builds"
	comparison "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/comparison"
	fetchcommit "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/fetchcommit"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/ignorechanges"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/planinstallgroup"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/postdeployrunbooks"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/previewimpact"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/setuppreview"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/updateinstallgroup"
)

const (
	ignoreChangesStepVersion  = "app-branch-ignore-changes-step-v1"
	setupPreviewHiddenVersion = "app-branch-setup-preview-hidden-v1"
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
	syncAppConfig := generics.FromPtrStr(flw.Metadata["sync_app_config"]) == "true"

	// Read the run rather than the workflow metadata: only the VCS push path
	// writes run_type into metadata, so a plan-only run triggered through the
	// API looked like a full deploy here.
	run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, runID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch app branch run")
	}
	isPreview := run.IsPreview()

	steps := make([]*app.WorkflowStep, 0)
	sg := newStepGroup()

	var changedFiles []string
	if raw := generics.FromPtrStr(flw.Metadata["changed_files"]); raw != "" {
		if err := json.Unmarshal([]byte(raw), &changedFiles); err != nil {
			return nil, errors.Wrap(err, "unable to decode changed files")
		}
	}

	if workflow.GetVersion(ctx, ignoreChangesStepVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion {
		sg.nextGroup()
		ignoreStep, err := sg.appBranchSignalStep(ctx, appBranchID, "check ignored changes", pgtype.Hstore{}, &ignorechanges.Signal{
			RunID:        runID,
			AppBranchID:  appBranchID,
			BaseSHA:      generics.FromPtrStr(flw.Metadata["base_sha"]),
			ChangedFiles: changedFiles,
		}, WithSkippable(false), WithExecutionType(app.WorkflowStepExecutionTypeHidden))
		if err != nil {
			return nil, errors.Wrap(err, "unable to create ignored changes step")
		}
		steps = append(steps, ignoreStep)
	}

	if isPreview {
		sg.nextGroup()
		options := []WorkflowStepOptions{WithSkippable(false)}
		if workflow.GetVersion(ctx, setupPreviewHiddenVersion, workflow.DefaultVersion, 1) != workflow.DefaultVersion {
			options = append(options, WithExecutionType(app.WorkflowStepExecutionTypeHidden))
		}
		step, err := sg.appBranchSignalStep(ctx, appBranchID, "setup preview", pgtype.Hstore{}, &setuppreview.Signal{
			RunID:       runID,
			AppBranchID: appBranchID,
		}, options...)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create setup preview step")
		}
		steps = append(steps, step)
	}

	switch {
	case appConfigID == "":
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

	case syncAppConfig:
		// Config was compiled by the caller: nothing to fetch, but it still has
		// to be synced into database records before the builds step can run.
		sg.nextGroup()
		step, err := sg.appBranchSignalStep(ctx, appBranchID, "fetch commit (skipped)", pgtype.Hstore{}, nil)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create skipped fetch commit step")
		}
		steps = append(steps, step)

		sg.nextGroup()
		step, err = sg.appBranchSignalStep(ctx, appBranchID, "sync app config", pgtype.Hstore{}, &appconfig.Signal{
			AppBranchID: appBranchID,
			RunID:       runID,
			AppConfigID: appConfigID,
		}, WithSkippable(false))
		if err != nil {
			return nil, errors.Wrap(err, "unable to create sync app config step")
		}
		steps = append(steps, step)

	default:
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

	sg.nextGroup()
	step, err := sg.appBranchSignalStep(ctx, appBranchID, "compute differences", pgtype.Hstore{}, &comparison.Signal{
		AppBranchID: appBranchID,
		RunID:       runID,
	}, WithSkippable(false))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create compute differences step")
	}
	steps = append(steps, step)

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

	// Preview runs: synthetic single-install group instead of previewimpact.
	if isPreview {
		if run.Preview != nil && run.Preview.InstallID != "" {
			switch run.Preview.Mode {
			case app.AppBranchRunPreviewModeBuildOnly:
				return sg.Result(steps), nil
			case app.AppBranchRunPreviewModePlanOnly:
				sg.nextGroup()
				step, err := sg.appBranchSignalStep(ctx, appBranchID, "plan preview install", pgtype.Hstore{}, &planinstallgroup.Signal{
					PreviewInstallID:   run.Preview.InstallID,
					SyntheticGroupName: "preview",
					AppBranchID:        appBranchID,
					RunID:              runID,
				}, WithSkippable(true))
				if err != nil {
					return nil, errors.Wrap(err, "unable to create preview plan step")
				}
				steps = append(steps, step)
			case app.AppBranchRunPreviewModeApply:
				sg.nextGroup()
				step, err := sg.appBranchSignalStep(ctx, appBranchID, "apply preview install", pgtype.Hstore{}, &updateinstallgroup.Signal{
					PreviewInstallID:   run.Preview.InstallID,
					SyntheticGroupName: "preview",
					AppBranchID:        appBranchID,
					RunID:              runID,
				}, WithSkippable(true))
				if err != nil {
					return nil, errors.Wrap(err, "unable to create preview apply step")
				}
				steps = append(steps, step)
			}
			return sg.Result(steps), nil
		}

		sg.nextGroup()
		step, err := sg.appBranchSignalStep(ctx, appBranchID, "preview install impact", pgtype.Hstore{}, &previewimpact.Signal{
			RunID:             runID,
			AppBranchID:       appBranchID,
			AppBranchConfigID: configID,
		}, WithSkippable(true))
		if err != nil {
			return nil, errors.Wrap(err, "unable to create preview impact step")
		}
		steps = append(steps, step)

		return sg.Result(steps), nil
	}

	allInstallGroups, err := activities.AwaitGetInstallGroupsByConfigID(ctx, configID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch install groups")
	}

	branchConfig, err := activities.AwaitGetAppBranchConfigByID(ctx, &activities.GetAppBranchConfigByIDInput{
		AppBranchConfigID: configID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to fetch app branch config")
	}
	hasPostDeployRunbooks := len(branchConfig.PostDeployRunbookIDs) > 0

	for _, group := range allInstallGroups {
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
		}, WithSkippable(true))
		if err != nil {
			return nil, errors.Wrapf(err, "unable to create deploy step for group %s", group.Name)
		}
		steps = append(steps, deployStep)

		// Only branches that configure post-deploy runbooks get the extra step, so
		// runs that don't use the feature keep the step list they had before.
		if hasPostDeployRunbooks {
			sg.nextGroup()
			runbooksStep, err := sg.appBranchSignalStep(ctx, appBranchID, "run post-deploy runbooks: "+group.Name, pgtype.Hstore{}, &postdeployrunbooks.Signal{
				InstallGroupID:    group.ID,
				AppBranchID:       appBranchID,
				RunID:             runID,
				AppBranchConfigID: configID,
			}, WithSkippable(true))
			if err != nil {
				return nil, errors.Wrapf(err, "unable to create post-deploy runbooks step for group %s", group.Name)
			}
			steps = append(steps, runbooksStep)
		}
	}

	return sg.Result(steps), nil
}
