package v2

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitrunnerhealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generatestate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/reprovisionrunner"
	statepartialgenerate "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/state/statepartialgenerate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

// ReprovisionStack recreates the install stack — and with it the runner — without
// touching the sandbox. Components are only redeployed when the caller opts in by
// leaving the skip_components metadata off; a bare stack reprovision leaves whatever
// is running on the sandbox alone.
//
// No pre/post reprovision lifecycle actions run around the stack itself: actions
// execute on the runner, and the runner is torn down and recreated mid-workflow, so a
// pre-hook would run against the old runner and a post-hook against the new one. The
// component deploys, when they run, still carry their own
// pre/post-deploy-all-components hooks.
func ReprovisionStack(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	steps := make([]*app.WorkflowStep, 0)
	sg := newStepGroup(flw)

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	stackSteps, err := getStackReprovisionSteps(ctx, sg, install, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, stackSteps...)

	if generics.FromPtrStr(flw.Metadata["skip_components"]) == "true" {
		return sg.Result(steps), nil
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	awData, err := activities.AwaitGetActionWorkflows(ctx, &activities.GetActionWorkflows{
		InstallID: installID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get action workflows")
	}

	dg := newGenCtx(sg, flw, installID, appCfg, awData, WithInstallInputs(install.CurrentInstallInputs))

	deploySteps, err := deployAllComponents(ctx, dg)
	if err != nil {
		return nil, err
	}
	steps = append(steps, deploySteps...)

	return sg.Result(steps), nil
}

// getStackReprovisionSteps emits the steps that recreate an install's stack: a new
// runner service account, a new stack version and the wait for its run, a state
// regeneration, and the wait for the new runner to report healthy. Shared by the
// stack-only reprovision and the full install reprovision.
func getStackReprovisionSteps(ctx workflow.Context, sg *stepGroup, install *app.Install, planOnly bool) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0)
	installID := install.ID

	sg.nextGroupEager() // reprovision service account
	step, err := sg.installSignalStep(ctx, installID, "reprovision runner service account", pgtype.Hstore{}, &reprovisionrunner.Signal{
		InstallID: installID,
	}, planOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	versionSteps, err := getStackVersionSteps(ctx, sg, installID, planOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, versionSteps...)

	sg.nextGroupEager() // generate install state (after stack is ready)
	orgEnabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureStateGenV2))
	if err != nil {
		return nil, errors.Wrap(err, "unable to check state-gen-v2 feature")
	}
	stateGenV2 := statemanager.UseStateGenV2(orgEnabled, install.Metadata)
	var stateSignal signal.Signal
	if stateGenV2 {
		stateSignal = &statepartialgenerate.Signal{
			InstallID:       installID,
			Targets:         statemanager.TargetsForHint(statemanager.HintInstallCreated, ""),
			TriggeredByID:   installID,
			TriggeredByType: "installs",
		}
	} else {
		stateSignal = &generatestate.Signal{InstallID: installID}
	}
	step, err = sg.installSignalStep(
		ctx,
		installID,
		"generate install state",
		pgtype.Hstore{},
		stateSignal,
		planOnly,
		WithSkippable(false),
	)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = sg.installSignalStep(ctx, installID, runnerHealthyStepName, pgtype.Hstore{}, &awaitrunnerhealthy.Signal{
		InstallID: installID,
	}, planOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	return steps, nil
}
