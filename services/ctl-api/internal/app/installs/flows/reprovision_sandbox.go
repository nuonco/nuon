package flows

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

func (w *Flows) ReprovisionSandbox(ctx workflow.Context, flw *app.Flow) ([]*app.FlowStep, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	steps := make([]*app.FlowStep, 0)

	step, err := w.installSignalStep(ctx, installID, "await runner health", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationAwaitRunnerHealthy,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err := w.getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreReprovisionSandbox)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	step, err = w.installSignalStep(ctx, installID, "reprovision sandbox", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationReprovisionSandbox,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = w.installSignalStep(ctx, installID, "sync secrets", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationSyncSecrets,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = w.installSignalStep(ctx, installID, "reprovision sandbox dns if enabled", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationProvisionDNS,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err = w.getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostReprovisionSandbox)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	deploySteps, err := w.deployAllComponents(ctx, installID, flw)
	if err != nil {
		return nil, err
	}
	steps = append(steps, deploySteps...)

	return steps, nil
}
