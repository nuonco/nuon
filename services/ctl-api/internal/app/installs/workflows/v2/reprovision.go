package v2

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/provisiondns"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/reprovisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/reprovisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/syncsecrets"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func Reprovision(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	steps := make([]*app.WorkflowStep, 0)

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	sg := newStepGroup(flw)

	stackSteps, err := getStackReprovisionSteps(ctx, sg, install, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, stackSteps...)

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

	dg := newGenCtx(sg, flw, installID, appCfg, awData, WithInstallInputs(install.CurrentInstallInputs))

	lifecycleSteps, err := getLifecycleActionsSteps(ctx, dg, app.ActionWorkflowTriggerTypePreReprovision)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	sandbox, err := activities.AwaitGetInstallSandboxByInstallID(ctx, installID)
	if err != nil {
		return nil, err
	}

	sg.nextGroup() // reprovision sandbox plan + apply

	step, err := sg.installSignalStep(ctx, installID, "reprovision sandbox plan", pgtype.Hstore{}, &reprovisionsandboxplan.Signal{
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
		}, flw.PlanOnly, WithMaxAutoRetries(install.AppSandboxConfig.GetMaxAutoRetries()))
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)

		lifecycleSteps, err = getLifecycleActionsSteps(ctx, dg, app.ActionWorkflowTriggerTypePreSecretsSync)
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

		lifecycleSteps, err = getLifecycleActionsSteps(ctx, dg, app.ActionWorkflowTriggerTypePostSecretsSync)
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

		deploySteps, err := deployAllComponents(ctx, dg, true)
		if err != nil {
			return nil, err
		}
		steps = append(steps, deploySteps...)

		lifecycleSteps, err = getLifecycleActionsSteps(ctx, dg, app.ActionWorkflowTriggerTypePostReprovision)
		if err != nil {
			return nil, err
		}
		steps = append(steps, lifecycleSteps...)
	}

	return sg.Result(steps), nil
}
