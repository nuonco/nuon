package flows

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

func (w *Flows) Deprovision(ctx workflow.Context, flw *app.Flow) ([]*app.FlowStep, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	steps := make([]*app.FlowStep, 0)

	step, err := w.installSignalStep(ctx, installID, "await runner healthy", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationAwaitRunnerHealthy,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	deploySteps, err := w.TeardownComponents(ctx, flw)
	if err != nil {
		return nil, err
	}
	steps = append(steps, deploySteps...)

	step, err = w.installSignalStep(ctx, installID, "deprovision sandbox", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationDeprovisionSandbox,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	return steps, nil
}
