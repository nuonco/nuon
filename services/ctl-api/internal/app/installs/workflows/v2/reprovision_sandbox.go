package v2

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/awaitrunnerhealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/generatestate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/provisiondns"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/reprovisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/reprovisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/syncsecrets"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

func ReprovisionSandbox(ctx workflow.Context, flw *app.Workflow) ([]*app.WorkflowStep, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	steps := make([]*app.WorkflowStep, 0)
	sg := newStepGroup()

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

	sandboxReprovisionSteps, err := getSandboxReprovisionSteps(ctx, installID, flw, sg, appCfg, awData)
	if err != nil {
		return nil, err
	}
	steps = append(steps, sandboxReprovisionSteps...)

	return steps, nil
}

func getSandboxReprovisionSteps(ctx workflow.Context, installID string, flw *app.Workflow, sg *stepGroup, appCfg *app.AppConfig, awData []*app.InstallActionWorkflow) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0)

	sg.nextGroup() // generate install state
	step, err := sg.installSignalStep(ctx, installID, "generate install state", pgtype.Hstore{}, &generatestate.Signal{
		InstallID: installID,
	}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	sg.nextGroup() // runner health
	step, err = sg.installSignalStep(ctx, installID, "await runner health", pgtype.Hstore{}, &awaitrunnerhealthy.Signal{
		InstallID: installID,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err := getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreReprovisionSandbox, sg, appCfg, awData)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	sandbox, err := activities.AwaitGetInstallSandboxByInstallID(ctx, installID)
	if err != nil {
		return nil, err
	}

	sg.nextGroup() // sandbox plan + apply
	step, err = sg.installSignalStep(ctx, installID, "reprovision sandbox plan", pgtype.Hstore{}, &reprovisionsandboxplan.Signal{
		InstallSandboxID: sandbox.ID,
		Role:             flw.Role,
	}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	if flw.PlanOnly {
		return steps, nil
	}

	step, err = sg.installSignalStep(ctx, installID, "reprovision sandbox apply", pgtype.Hstore{}, &reprovisionsandboxapplyplan.Signal{
		InstallSandboxID: sandbox.ID,
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

	sg.nextGroup()
	step, err = sg.installSignalStep(ctx, installID, "reprovision sandbox dns if enabled", pgtype.Hstore{}, &provisiondns.Signal{
		InstallID: installID,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	if generics.FromPtrStr(flw.Metadata["skip_components"]) != "true" {
		deploySteps, err := deployAllComponents(ctx, installID, flw, sg, appCfg, awData)
		if err != nil {
			return nil, err
		}
		steps = append(steps, deploySteps...)
	}

	lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostReprovisionSandbox, sg, appCfg, awData)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	return steps, nil
}
