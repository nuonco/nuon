package workflows

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

func DeprovisionSandbox(ctx workflow.Context, flw *app.Workflow) ([]*app.WorkflowStep, error) {
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

	sg.nextGroup() // runner health
	step, err = sg.installSignalStep(ctx, installID, "await runner healthy", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationAwaitRunnerHealthy,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err := getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreDeprovisionSandbox, sg)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	sg.nextGroup() // deprovision sandbox plan + apply

	step, err = sg.installSignalStep(ctx, installID, "deprovision sandbox plan", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationDeprovisionSandboxPlan,
	}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = sg.installSignalStep(ctx, installID, "deprovision sandbox apply plan", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationDeprovisionSandboxApplyPlan,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostDeprovisionSandbox, sg)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	return steps, nil
}
