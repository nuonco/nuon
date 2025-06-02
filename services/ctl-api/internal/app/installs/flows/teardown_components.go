package flows

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.temporal.io/sdk/workflow"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func (w *Flows) TeardownComponents(ctx workflow.Context, flw *app.Flow) ([]*app.FlowStep, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	componentIDs, err := activities.AwaitGetAppGraph(ctx, activities.GetAppGraphRequest{
		InstallID: install.ID,
		Reverse:   true,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install graph")
	}

	steps := make([]*app.FlowStep, 0)
	for _, compID := range componentIDs {
		comp, err := activities.AwaitGetComponentByComponentID(ctx, compID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get component")
		}
		installComp, err := activities.AwaitGetInstallComponent(ctx, activities.GetInstallComponentRequest{
			InstallID:   installID,
			ComponentID: comp.ID,
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to get install component")
		}

		if installComp == nil {
			continue
		}

		if installComp.Status == app.InstallComponentStatusInactive {
			reason := fmt.Sprintf("install component %s is inactive", comp.Name)
			deployStep, err := w.installSignalStep(ctx, installID, "skipped teardown "+comp.Name, pgtype.Hstore{
				"reason": &reason,
			}, nil)
			if err != nil {
				return nil, errors.Wrap(err, "unable to create skip step")
			}
			steps = append(steps, deployStep)

			continue
		}

		preDeploySteps, err := w.getComponentLifecycleActionsSteps(ctx, flw.ID, compID, installID, app.ActionWorkflowTriggerTypePreTeardownComponent)
		if err != nil {
			return nil, err
		}
		steps = append(steps, preDeploySteps...)

		deployStep, err := w.installSignalStep(ctx, installID, "teardown "+comp.Name, pgtype.Hstore{}, &signals.Signal{
			Type: signals.OperationExecuteTeardownComponent,
			ExecuteTeardownComponentSubSignal: signals.TeardownComponentSubSignal{
				ComponentID: compID,
			},
		})
		steps = append(steps, deployStep)

		postDeploySteps, err := w.getComponentLifecycleActionsSteps(ctx, flw.ID, compID, installID, app.ActionWorkflowTriggerTypePostTeardownComponent)
		if err != nil {
			return nil, err
		}
		steps = append(steps, postDeploySteps...)
	}

	return steps, nil
}
