package v2

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitinstallstackversionrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitrunnerhealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generateinstallstackversion"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/provisiondns"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/provisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/provisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/syncsecrets"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow"
)

func Provision(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	steps := make([]*app.WorkflowStep, 0)

	sg := newStepGroup(flw)

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	// Resolve stack ID for install stack signals
	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, installID)
	if err != nil {
		return nil, err
	}
	stackID := stack.ID

	if flw.PlanOnly {
		// Plan-only keeps the three separate steps: getSignalStepMetadata skips
		// the stack signal here but not state generation, so folding them into
		// one step would stop generating state on a plan-only run.
		planOnlySteps, err := provisionStatePlanOnlySteps(ctx, sg, flw, install, installID, stackID)
		if err != nil {
			return nil, err
		}
		steps = append(steps, planOnlySteps...)
	} else {
		sg.nextGroupEager()

		step, err := sg.installSignalStep(ctx, installID, "generate install stack", pgtype.Hstore{}, &generateinstallstackversion.Signal{
			InstallStackID:        stackID,
			PrepareStateAndRunner: true,
		}, flw.PlanOnly, WithSkippable(false), WithRetryable(true))
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	sg.nextGroupEager() // await install stack

	step, err := sg.installSignalStep(ctx, installID, "await install stack", pgtype.Hstore{}, &awaitinstallstackversionrun.Signal{
		InstallStackID:     stackID,
		CreateManagedStack: true,
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

	// Everything above is eager, and everything below needs the app config,
	// action workflows, sandbox and component graph — reads that cost far more
	// than the steps generated so far. Release the eager groups here so the
	// conductor executes them while the rest of this generator runs, instead of
	// idling until it returns.
	flow.PublishEagerGroups(ctx, sg.Groups(), steps)

	// Stop with the stack and runner up so inputs that are only knowable once the
	// runner exists can be set before the sandbox reads them. Provisioning the
	// sandbox and components is a separate, explicit step afterwards.
	if flw.IsStackOnly() {
		return sg.Result(steps), nil
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

	dg := newGenCtx(sg, flw, installID, appCfg, awData, WithInstallInputs(install.CurrentInstallInputs))

	lifecycleSteps, err := getLifecycleActionsSteps(ctx, dg, app.ActionWorkflowTriggerTypePreProvision)
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
		}, flw.PlanOnly, WithMaxAutoRetries(install.AppSandboxConfig.GetMaxAutoRetries()))
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)

		lifecycleSteps, err = getLifecycleActionsSteps(ctx, dg, app.ActionWorkflowTriggerTypePostProvisionSandbox)
		if err != nil {
			return nil, err
		}
		steps = append(steps, lifecycleSteps...)

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

		sg.nextGroup() // provision sandbox dns
		step, err = sg.installSignalStep(ctx, installID, "provision sandbox dns if enabled", pgtype.Hstore{}, &provisiondns.Signal{
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

		lifecycleSteps, err = getLifecycleActionsSteps(ctx, dg, app.ActionWorkflowTriggerTypePostProvision)
		if err != nil {
			return nil, err
		}
		steps = append(steps, lifecycleSteps...)
	}

	return sg.Result(steps), nil
}
