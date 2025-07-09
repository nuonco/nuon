package flows

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

func Deprovision(ctx workflow.Context, flw *app.Workflow) ([]*app.WorkflowStep, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	steps := make([]*app.WorkflowStep, 0)

	step, err := installSignalStep(ctx, installID, "await runner healthy", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationAwaitRunnerHealthy,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err := getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreDeprovision)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	deploySteps, err := TeardownComponents(ctx, flw)
	if err != nil {
		return nil, err
	}
	steps = append(steps, deploySteps...)

	step, err = installSignalStep(ctx, installID, "deprovision sandbox plan", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationDeprovisionSandboxPlan,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = installSignalStep(ctx, installID, "deprovision sandbox apply", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationDeprovisionSandboxApplyPlan,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostDeprovision)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	return steps, nil
}
