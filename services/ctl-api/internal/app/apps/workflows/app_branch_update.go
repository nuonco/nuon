package workflows

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/updateinstallgroup"
)

// AppBranchUpdate generates workflow steps for updating specific installs to a
// new app config without running the full branch pipeline (no build step).
// Used for re-deploying individual installs or install groups.
//
// Required metadata:
//   - app_branch_id: the app branch
//   - run_id: the app branch run
//   - config_id: the app branch config
//   - app_config_id: the target app config to deploy
//
// Optional metadata:
//   - install_id: update a single install (mutually exclusive with install_group_id)
//   - install_group_id: update a specific install group
func AppBranchUpdate(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	appBranchID := generics.FromPtrStr(flw.Metadata["app_branch_id"])
	if appBranchID == "" {
		return nil, errors.New("app_branch_id not found in workflow metadata")
	}

	runID := generics.FromPtrStr(flw.Metadata["run_id"])
	if runID == "" {
		return nil, errors.New("run_id not found in workflow metadata")
	}

	appConfigID := generics.FromPtrStr(flw.Metadata["app_config_id"])
	if appConfigID == "" {
		return nil, errors.New("app_config_id not found in workflow metadata")
	}

	installGroupID := generics.FromPtrStr(flw.Metadata["install_group_id"])
	installID := generics.FromPtrStr(flw.Metadata["install_id"])

	steps := make([]*app.WorkflowStep, 0)
	sg := newStepGroup()

	if installID != "" {
		// Single install update: create one install config update workflow
		sg.nextGroup()
		step, err := sg.appBranchSignalStep(ctx, appBranchID, "update install", pgtype.Hstore{}, &updateinstallgroup.Signal{
			InstallGroupID: installGroupID,
			AppBranchID:    appBranchID,
			RunID:          runID,
		}, WithSkippable(false))
		if err != nil {
			return nil, errors.Wrap(err, "unable to create update step")
		}
		steps = append(steps, step)
	} else if installGroupID != "" {
		// Single group update
		sg.nextGroup()
		step, err := sg.appBranchSignalStep(ctx, appBranchID, "deploy install group", pgtype.Hstore{}, &updateinstallgroup.Signal{
			InstallGroupID: installGroupID,
			AppBranchID:    appBranchID,
			RunID:          runID,
		}, WithSkippable(false))
		if err != nil {
			return nil, errors.Wrap(err, "unable to create deploy group step")
		}
		steps = append(steps, step)
	} else {
		// Update all install groups (like AppBranchRun but without builds)
		configID := generics.FromPtrStr(flw.Metadata["config_id"])
		if configID == "" {
			return nil, errors.New("config_id not found in workflow metadata")
		}

		installGroups, err := activities.AwaitGetInstallGroupsByConfigID(ctx, configID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to fetch install groups")
		}

		for _, group := range installGroups {
			sg.nextGroup()
			step, err := sg.appBranchSignalStep(ctx, appBranchID, "deploy install group: "+group.Name, pgtype.Hstore{}, &updateinstallgroup.Signal{
				InstallGroupID: group.ID,
				AppBranchID:    appBranchID,
				RunID:          runID,
			}, WithSkippable(false))
			if err != nil {
				return nil, errors.Wrapf(err, "unable to create deploy step for group %s", group.Name)
			}
			steps = append(steps, step)
		}
	}

	return sg.Result(steps), nil
}
