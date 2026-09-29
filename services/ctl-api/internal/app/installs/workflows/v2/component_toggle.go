package v2

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	statepartialgenerate "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/state/statepartialgenerate"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

func stateInputsRefreshStep(ctx workflow.Context, sg *stepGroup, install *app.Install, planOnly bool) (*app.WorkflowStep, error) {
	stateSignal := &statepartialgenerate.Signal{
		InstallID:       install.ID,
		Targets:         statemanager.TargetsForHint(statemanager.HintInputsUpdated, ""),
		TriggeredByID:   install.ID,
		TriggeredByType: "installs",
	}

	return sg.installSignalStep(
		ctx,
		install.ID,
		"update install state inputs",
		pgtype.Hstore{},
		stateSignal, planOnly,
		WithSkippable(false),
	)
}

func ComponentEnabledSteps(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	return InputUpdate(ctx, flw)
}

func ComponentDisabledSteps(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	return InputUpdate(ctx, flw)
}
