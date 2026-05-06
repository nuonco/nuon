package v2

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/awaitinstallstackversionrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/awaitrunnerhealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/generateinstallstackversion"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/generatestate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/provisiondns"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/reprovisionrunner"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/reprovisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/reprovisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/syncsecrets"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/updateinstallstackoutputs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func Reprovision(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	steps := make([]*app.WorkflowStep, 0)

	sg := newStepGroup(flw)

	sg.nextGroupEager() // reprovision service account
	step, err := sg.installSignalStep(ctx, installID, "reprovision runner service account", pgtype.Hstore{}, &reprovisionrunner.Signal{
		InstallID: installID,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, installID)
	if err != nil {
		return nil, err
	}

	sg.nextGroupEager() // install stack

	step, err = sg.installSignalStep(ctx, installID, "generate install stack", pgtype.Hstore{}, &generateinstallstackversion.Signal{
		InstallStackID: stack.ID,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = sg.installSignalStep(ctx, installID, "await install stack", pgtype.Hstore{}, &awaitinstallstackversionrun.Signal{
		InstallStackID: stack.ID,
	}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = sg.installSignalStep(ctx, installID, "update install stack outputs", pgtype.Hstore{}, &updateinstallstackoutputs.Signal{
		InstallStackID:          stack.ID,
		SkipInputUpdateWorkflow: true,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	sg.nextGroupEager() // generate install state (after stack is ready)
	step, err = sg.installSignalStep(ctx, installID, "generate install state", pgtype.Hstore{}, &generatestate.Signal{
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

	lifecycleSteps, err := getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreReprovision, sg, appCfg, awData)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	sandbox, err := activities.AwaitGetInstallSandboxByInstallID(ctx, installID)
	if err != nil {
		return nil, err
	}

	sg.nextGroup() // reprovision sandbox plan + apply

	step, err = sg.installSignalStep(ctx, installID, "reprovision sandbox plan", pgtype.Hstore{}, &reprovisionsandboxplan.Signal{
		InstallSandboxID: sandbox.ID,
		InstallID:        installID,
		Role:             flw.Role,
	}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	if !flw.PlanOnly {
		step, err = sg.installSignalStep(ctx, installID, "reprovision sandbox apply plan", pgtype.Hstore{}, &reprovisionsandboxapplyplan.Signal{
			InstallSandboxID: sandbox.ID,
			InstallID:        installID,
		}, flw.PlanOnly)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)

		lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreSecretsSync, sg, appCfg, awData)
		if err != nil {
			return nil, err
		}
		steps = append(steps, lifecycleSteps...)

		sg.nextGroup() // sync secrets
		step, err = sg.installSignalStep(ctx, installID, "sync secrets", pgtype.Hstore{}, &syncsecrets.Signal{
			InstallID: installID,
		}, flw.PlanOnly, WithSkippable(false))
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)

		lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostSecretsSync, sg, appCfg, awData)
		if err != nil {
			return nil, err
		}
		steps = append(steps, lifecycleSteps...)

		sg.nextGroup() // reprovision sandbox dns
		step, err = sg.installSignalStep(ctx, installID, "reprovision sandbox dns if enabled", pgtype.Hstore{}, &provisiondns.Signal{
			InstallID: installID,
		}, flw.PlanOnly)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)

		deploySteps, err := deployAllComponents(ctx, installID, flw, sg, appCfg, awData)
		if err != nil {
			return nil, err
		}
		steps = append(steps, deploySteps...)

		lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostReprovision, sg, appCfg, awData)
		if err != nil {
			return nil, err
		}
		steps = append(steps, lifecycleSteps...)
	}

	return sg.Result(steps), nil
}
