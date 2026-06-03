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

	sandboxReprovisionSteps, err := getSandboxReprovisionSteps(ctx, install, installID, flw, sg, appCfg, awData)
	if err != nil {
		return nil, err
	}
	steps = append(steps, sandboxReprovisionSteps...)

	return sg.Result(steps), nil
}

func getSandboxReprovisionSteps(ctx workflow.Context, install *app.Install, installID string, flw *app.Workflow, sg *stepGroup, appCfg *app.AppConfig, awData []*app.InstallActionWorkflow) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0)

	sg.nextGroupEager() // generate install state
	orgEnabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureStateGenV2))
	if err != nil {
		return nil, errors.Wrap(err, "unable to check state-gen-v2 feature")
	}
	stateGenV2 := statemanager.UseStateGenV2(orgEnabled, install.Metadata)
	if !stateGenV2 {
		stateSignal := &generatestate.Signal{InstallID: installID}
		step, err := sg.installSignalStep(ctx, installID, "generate install state", pgtype.Hstore{}, stateSignal, flw.PlanOnly, WithSkippable(false))
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	step, err := sg.installSignalStep(ctx, installID, "runner healthy", pgtype.Hstore{}, &awaitrunnerhealthy.Signal{
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
		InstallID:        installID,
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
		InstallID:        installID,
	}, flw.PlanOnly, WithMaxAutoRetries(install.AppSandboxConfig.GetMaxAutoRetries()))
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
