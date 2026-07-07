package v2

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generatestate"
	statepartialgenerate "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/state/statepartialgenerate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

// stateInputsRefreshStep builds the "update install state inputs" step that
// refreshes the inputs partial in install state after an input change. Shared
// by the input-update workflow and the component enable/disable workflows,
// which are all driven by a synthetic-input change. The caller is responsible
// for opening the step group (sg.nextGroup()) beforehand.
func stateInputsRefreshStep(ctx workflow.Context, sg *stepGroup, install *app.Install, planOnly bool) (*app.WorkflowStep, error) {
	orgEnabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureStateGenV2))
	if err != nil {
		return nil, errors.Wrap(err, "unable to check state-gen-v2 feature")
	}
	stateGenV2 := statemanager.UseStateGenV2(orgEnabled, install.Metadata)

	var stateSignal signal.Signal
	if stateGenV2 {
		stateSignal = &statepartialgenerate.Signal{
			InstallID:       install.ID,
			Targets:         statemanager.TargetsForHint(statemanager.HintInputsUpdated, ""),
			TriggeredByID:   install.ID,
			TriggeredByType: "installs",
		}
	} else {
		stateSignal = &generatestate.Signal{InstallID: install.ID}
	}

	return sg.installSignalStep(ctx, install.ID, "update install state inputs", pgtype.Hstore{}, stateSignal, planOnly, WithSkippable(false))
}

// ComponentEnabledSteps and ComponentDisabledSteps back the dedicated
// WorkflowTypeComponentEnabled / WorkflowTypeComponentDisabled workflows. These
// types exist purely for UX and observability: a toggle gets its own
// recognizable workflow name ("Enabling component" / "Disabling component")
// instead of a generic input update.
//
// They deliberately carry no behavior of their own. The synthetic enabled input
// is the single source of truth and is persisted by the service layer before
// the workflow starts, so reconciliation is identical to a config-file toggle.
// Delegating straight to InputUpdate guarantees the dedicated path can never
// drift from the generic input-update path.
func ComponentEnabledSteps(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	return InputUpdate(ctx, flw)
}

func ComponentDisabledSteps(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	return InputUpdate(ctx, flw)
}
