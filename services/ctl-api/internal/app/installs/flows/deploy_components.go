package flows

import (
	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.temporal.io/sdk/workflow"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func DeployAllComponents(ctx workflow.Context, flw *app.Flow) ([]*app.FlowStep, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	componentIDs, err := activities.AwaitGetAppGraph(ctx, activities.GetAppGraphRequest{
		InstallID: install.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install graph")
	}

	steps := make([]*app.FlowStep, 0)
	step, err := installSignalStep(ctx, installID, "await runner healthy", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationAwaitRunnerHealthy,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err := getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreDeployAllComponents)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)
	deploySteps, err := getComponentDeploySteps(ctx, installID, flw, componentIDs)
	if err != nil {
		return nil, err
	}

	steps = append(steps, deploySteps...)

	lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostDeployAllComponents)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)
	return steps, nil
}
