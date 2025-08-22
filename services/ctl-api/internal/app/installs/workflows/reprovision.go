package workflows

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

func Reprovision(ctx workflow.Context, flw *app.Workflow) ([]*app.WorkflowStep, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	steps := make([]*app.WorkflowStep, 0)

	sg := newStepGroup()

	sg.nextGroup() // generate install state
	step, err := sg.installSignalStep(ctx, installID, "generate install state", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationGenerateState,
	}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}

	steps = append(steps, step)

	sg.nextGroup() // reprovision service account
	step, err = sg.installSignalStep(ctx, installID, "reprovision runner service account", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationReprovisionRunner,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	sg.nextGroup() // install stack

	step, err = sg.installSignalStep(ctx, installID, "generate install stack", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationGenerateInstallStackVersion,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = sg.installSignalStep(ctx, installID, "await install stack", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationAwaitInstallStackVersionRun,
	}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = sg.installSignalStep(ctx, installID, "update install stack outputs", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationUpdateInstallStackOutputs,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	sg.nextGroup() // await runner
	step, err = sg.installSignalStep(ctx, installID, "await runner health", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationAwaitRunnerHealthy,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err := getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreReprovision, sg)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	sg.nextGroup() // reprovisoins sandbox plan + apply

	step, err = sg.installSignalStep(ctx, installID, "reprovision sandbox plan", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationReprovisionSandboxPlan,
	}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	if !flw.PlanOnly {
		step, err = sg.installSignalStep(ctx, installID, "reprovision sandbox apply plan", pgtype.Hstore{}, &signals.Signal{
			Type: signals.OperationReprovisionSandboxApplyPlan,
		}, flw.PlanOnly)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)

		lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreSecretsSync, sg)
		if err != nil {
			return nil, err
		}
		steps = append(steps, lifecycleSteps...)

		sg.nextGroup() // sync secrets
		step, err = sg.installSignalStep(ctx, installID, "sync secrets", pgtype.Hstore{}, &signals.Signal{
			Type: signals.OperationSyncSecrets,
		}, flw.PlanOnly, WithSkippable(false))
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)

		lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostSecretsSync, sg)
		if err != nil {
			return nil, err
		}
		steps = append(steps, lifecycleSteps...)

		sg.nextGroup() // reprovision sandbox dns
		step, err = sg.installSignalStep(ctx, installID, "reprovision sandbox dns if enabled", pgtype.Hstore{}, &signals.Signal{
			Type: signals.OperationProvisionDNS,
		}, flw.PlanOnly)
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)

		deploySteps, err := deployAllComponents(ctx, installID, flw, sg)
		if err != nil {
			return nil, err
		}
		steps = append(steps, deploySteps...)

		lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostReprovision, sg)
		if err != nil {
			return nil, err
		}
		steps = append(steps, lifecycleSteps...)
	}

	return steps, nil
}
