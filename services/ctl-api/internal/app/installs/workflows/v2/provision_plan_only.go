package v2

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generateinstallstackversion"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generatestate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/provisionrunner"
	statepartialgenerate "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/state/statepartialgenerate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

// provisionStatePlanOnlySteps emits state generation, the runner service
// account and stack generation as three separate steps. A real provision folds
// all three into one signal, but plan-only cannot: getSignalStepMetadata marks
// the stack signal skipped and the other two not, so a single step would either
// skip state generation or perform a stack-version write during a plan.
func provisionStatePlanOnlySteps(
	ctx workflow.Context,
	sg *stepGroup,
	flw *app.Workflow,
	install *app.Install,
	installID string,
	stackID string,
) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0, 3)

	sg.nextGroupEager()

	orgEnabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureStateGenV2))
	if err != nil {
		return nil, errors.Wrap(err, "unable to check state-gen-v2 feature")
	}

	var stateSignal signal.Signal
	if statemanager.UseStateGenV2(orgEnabled, install.Metadata) {
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

	sg.nextGroupEagerParallel()

	step, err = sg.installSignalStep(ctx, installID, "provision runner service account", pgtype.Hstore{}, &provisionrunner.Signal{
		InstallID: installID,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = sg.installSignalStep(ctx, installID, "generate install stack", pgtype.Hstore{}, &generateinstallstackversion.Signal{
		InstallStackID: stackID,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	return steps, nil
}
