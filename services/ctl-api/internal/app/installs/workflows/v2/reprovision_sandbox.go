package v2

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitrunnerhealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generatestate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/provisiondns"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/reprovisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/reprovisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/syncsecrets"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

func ReprovisionSandbox(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
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

	dg := newGenCtx(sg, flw, installID, appCfg, awData)

	sandboxReprovisionSteps, err := getSandboxReprovisionSteps(ctx, dg, install)
	if err != nil {
		return nil, err
	}
	steps = append(steps, sandboxReprovisionSteps...)

	return sg.Result(steps), nil
}

func getSandboxReprovisionSteps(ctx workflow.Context, dg *genCtx, install *app.Install) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0)

	dg.sg.nextGroupEager() // generate install state
	orgEnabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureStateGenV2))
	if err != nil {
		return nil, errors.Wrap(err, "unable to check state-gen-v2 feature")
	}
	stateGenV2 := statemanager.UseStateGenV2(orgEnabled, install.Metadata)
	if !stateGenV2 {
		stateSignal := &generatestate.Signal{InstallID: dg.installID}
		step, err := dg.sg.installSignalStep(ctx, dg.installID, "generate install state", pgtype.Hstore{}, stateSignal, dg.flw.PlanOnly, WithSkippable(false))
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	step, err := dg.sg.installSignalStep(ctx, dg.installID, "runner healthy", pgtype.Hstore{}, &awaitrunnerhealthy.Signal{
		InstallID: dg.installID,
	}, dg.flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err := getLifecycleActionsSteps(ctx, dg, app.ActionWorkflowTriggerTypePreReprovisionSandbox)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	sandbox, err := activities.AwaitGetInstallSandboxByInstallID(ctx, dg.installID)
	if err != nil {
		return nil, err
	}

	dg.sg.nextGroup() // sandbox plan + apply
	step, err = dg.sg.installSignalStep(ctx, dg.installID, "reprovision sandbox plan", pgtype.Hstore{}, &reprovisionsandboxplan.Signal{
		InstallSandboxID: sandbox.ID,
		InstallID:        dg.installID,
		Role:             dg.flw.Role,
	}, dg.flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	if dg.flw.PlanOnly {
		return steps, nil
	}

	step, err = dg.sg.installSignalStep(ctx, dg.installID, "reprovision sandbox apply", pgtype.Hstore{}, &reprovisionsandboxapplyplan.Signal{
		InstallSandboxID: sandbox.ID,
		InstallID:        dg.installID,
	}, dg.flw.PlanOnly, WithMaxAutoRetries(install.AppSandboxConfig.GetMaxAutoRetries()))
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err = getLifecycleActionsSteps(ctx, dg, app.ActionWorkflowTriggerTypePreSecretsSync)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	dg.sg.nextGroup() // sync secrets
	step, err = dg.sg.installSignalStep(ctx, dg.installID, "sync secrets", pgtype.Hstore{}, &syncsecrets.Signal{
		InstallID: dg.installID,
	}, dg.flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err = getLifecycleActionsSteps(ctx, dg, app.ActionWorkflowTriggerTypePostSecretsSync)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	dg.sg.nextGroup()
	step, err = dg.sg.installSignalStep(ctx, dg.installID, "reprovision sandbox dns if enabled", pgtype.Hstore{}, &provisiondns.Signal{
		InstallID: dg.installID,
	}, dg.flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	if generics.FromPtrStr(dg.flw.Metadata["skip_components"]) != "true" {
		deploySteps, err := deployAllComponents(ctx, dg)
		if err != nil {
			return nil, err
		}
		steps = append(steps, deploySteps...)
	}

	lifecycleSteps, err = getLifecycleActionsSteps(ctx, dg, app.ActionWorkflowTriggerTypePostReprovisionSandbox)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	return steps, nil
}
