package workflows

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

func Provision(ctx workflow.Context, flw *app.Workflow) ([]*app.WorkflowStep, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	steps := make([]*app.WorkflowStep, 0)

	sg := newStepGroup()

	sg.nextGroup() // generate install state
	step, err := sg.installSignalStep(ctx, installID, "generate install state", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationGenerateState,
	}, flw.PlanOnly, WithSkippable(false))
	steps = append(steps, step)

	sg.nextGroup() // provision service accoun

	step, err = sg.installSignalStep(ctx, installID, "provision runner service account", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationProvisionRunner,
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

	sg.nextGroup() // runner health
	step, err = sg.installSignalStep(ctx, installID, "await runner health", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationAwaitRunnerHealthy,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err := getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreProvision, sg)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	sg.nextGroup() // provision sandbox plan + apply
	step, err = sg.installSignalStep(ctx, installID, "provision sandbox plan", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationProvisionSandboxPlan,
	}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = sg.installSignalStep(ctx, installID, "provision sandbox apply plan", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationProvisionSandboxApplyPlan,
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

	sg.nextGroup() // sync secret
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

	sg.nextGroup() // provision sandbox dns
	step, err = sg.installSignalStep(ctx, installID, "provision sandbox dns if enabled", pgtype.Hstore{}, &signals.Signal{
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

	lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostProvision, sg)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	return steps, nil
}
