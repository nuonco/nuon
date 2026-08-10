package workflows

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/installsync/dispatchsyncs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/installsync/fetchcommit"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/installsync/parseconfigs"
)

func AppInstallSync(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	appID := generics.FromPtrStr(flw.Metadata["app_id"])
	if appID == "" {
		return nil, errors.New("app_id not found in workflow metadata")
	}

	syncID := generics.FromPtrStr(flw.Metadata["sync_id"])
	if syncID == "" {
		return nil, errors.New("sync_id not found in workflow metadata")
	}

	commitSHA := generics.FromPtrStr(flw.Metadata["commit_sha"])
	triggeredBy := generics.FromPtrStr(flw.Metadata["triggered_by"])
	installsDir := generics.FromPtrStr(flw.Metadata["installs_directory"])

	steps := make([]*app.WorkflowStep, 0)
	sg := newStepGroup()

	sg.nextGroup()
	step, err := sg.appSignalStep(ctx, appID, "fetch commit", pgtype.Hstore{}, &fetchcommit.Signal{
		AppID:                  appID,
		AppInstallConfigSyncID: syncID,
		CommitSHA:              commitSHA,
		TriggeredBy:            triggeredBy,
	}, WithSkippable(false))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create fetch commit step")
	}
	steps = append(steps, step)

	sg.nextGroup()
	step, err = sg.appSignalStep(ctx, appID, "parse install configs", pgtype.Hstore{}, &parseconfigs.Signal{
		AppID:                  appID,
		AppInstallConfigSyncID: syncID,
		InstallsDirectory:      installsDir,
		CommitSHA:              commitSHA,
	}, WithSkippable(false), WithExecutionType(app.WorkflowStepExecutionTypeApproval))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create parse configs step")
	}
	steps = append(steps, step)

	sg.nextGroup()
	step, err = sg.appSignalStep(ctx, appID, "sync installs", pgtype.Hstore{}, &dispatchsyncs.Signal{
		AppID:                  appID,
		AppInstallConfigSyncID: syncID,
		InstallsDirectory:      installsDir,
		CommitSHA:              commitSHA,
		TriggeredBy:            triggeredBy,
	}, WithSkippable(false))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create dispatch syncs step")
	}
	steps = append(steps, step)

	return sg.Result(steps), nil
}
