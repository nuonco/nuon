package v2

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/awaitrunnerhealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/deprovisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/deprovisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/generatestate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func DeprovisionSandbox(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	steps := make([]*app.WorkflowStep, 0)
	sg := newStepGroup(flw)

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	awData, err := activities.AwaitGetActionWorkflows(ctx, &activities.GetActionWorkflows{
		InstallID: installID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get action workflows")
	}

	sg.nextGroupEager()
	step, err := sg.installSignalStep(ctx, installID, "generate install state", pgtype.Hstore{}, &generatestate.Signal{
		InstallID: installID,
	}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = sg.installSignalStep(ctx, installID, "runner healthy", pgtype.Hstore{}, &awaitrunnerhealthy.Signal{
		InstallID: installID,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err := getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreDeprovisionSandbox, sg, appCfg, awData)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	sandbox, err := activities.AwaitGetInstallSandboxByInstallID(ctx, installID)
	if err != nil {
		return nil, err
	}

	sg.nextGroup() // deprovision sandbox plan + apply

	step, err = sg.installSignalStep(ctx, installID, "deprovision sandbox plan", pgtype.Hstore{}, &deprovisionsandboxplan.Signal{
		InstallSandboxID: sandbox.ID,
		InstallID:        installID,
		Role:             flw.Role,
	}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = sg.installSignalStep(ctx, installID, "deprovision sandbox apply plan", pgtype.Hstore{}, &deprovisionsandboxapplyplan.Signal{
		InstallSandboxID: sandbox.ID,
		InstallID:        installID,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostDeprovisionSandbox, sg, appCfg, awData)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	return sg.Result(steps), nil
}
