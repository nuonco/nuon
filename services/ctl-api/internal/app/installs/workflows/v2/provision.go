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
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/provisionrunner"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/provisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/provisionsandboxplan"
	statepartialgenerate "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/state/statepartialgenerate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/syncsecrets"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/updateinstallstackoutputs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

func Provision(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	steps := make([]*app.WorkflowStep, 0)

	sg := newStepGroup(flw)

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	sg.nextGroupEager()
	orgEnabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureStateGenV2))
	if err != nil {
		return nil, errors.Wrap(err, "unable to check state-gen-v2 feature")
	}
	stateGenV2 := statemanager.UseStateGenV2(orgEnabled, install.Metadata)
	var stateSignal signal.Signal
	if stateGenV2 {
		stateSignal = &statepartialgenerate.Signal{
			InstallID:       installID,
			Targets:         statemanager.TargetsForHint(statemanager.HintInstallCreated, ""),
			TriggeredByID:   installID,
			TriggeredByType: "installs",
		}
	} else {
		stateSignal = &generatestate.Signal{InstallID: installID}
	}
	step, err := sg.installSignalStep(ctx, installID, "generate install state", pgtype.Hstore{}, stateSignal, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	sg.nextGroupEager() // provision service account

	step, err = sg.installSignalStep(ctx, installID, "provision runner service account", pgtype.Hstore{}, &provisionrunner.Signal{
		InstallID: installID,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	// Resolve stack ID for install stack signals
	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, installID)
	if err != nil {
		return nil, err
	}
	stackID := stack.ID

	sg.nextGroupEager() // install stack

	step, err = sg.installSignalStep(ctx, installID, "generate install stack", pgtype.Hstore{}, &generateinstallstackversion.Signal{
		InstallStackID: stackID,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = sg.installSignalStep(ctx, installID, "await install stack", pgtype.Hstore{}, &awaitinstallstackversionrun.Signal{
		InstallStackID: stackID,
	}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = sg.installSignalStep(ctx, installID, "update install stack outputs", pgtype.Hstore{}, &updateinstallstackoutputs.Signal{
		InstallStackID:          stackID,
		SkipInputUpdateWorkflow: true,
	}, flw.PlanOnly)
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

	lifecycleSteps, err := getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreProvision, sg, appCfg, awData)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	// Resolve sandbox ID for sandbox signals
	sandbox, err := activities.AwaitGetInstallSandboxByInstallID(ctx, installID)
	if err != nil {
		return nil, err
	}
	sandboxID := sandbox.ID

	sg.nextGroup() // provision sandbox plan + apply
	step, err = sg.installSignalStep(ctx, installID, "provision sandbox plan", pgtype.Hstore{}, &provisionsandboxplan.Signal{
		InstallSandboxID: sandboxID,
		InstallID:        installID,
		Role:             flw.Role,
	}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	// Only proceed with remaining steps if not in plan-only mode
	if !flw.PlanOnly {
		step, err = sg.installSignalStep(ctx, installID, "provision sandbox apply plan", pgtype.Hstore{}, &provisionsandboxapplyplan.Signal{
			InstallSandboxID: sandboxID,
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

		sg.nextGroup() // provision sandbox dns
		step, err = sg.installSignalStep(ctx, installID, "provision sandbox dns if enabled", pgtype.Hstore{}, &provisiondns.Signal{
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

		lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostProvision, sg, appCfg, awData)
		if err != nil {
			return nil, err
		}
		steps = append(steps, lifecycleSteps...)
	}

	return sg.Result(steps), nil
}
