package v2

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/actionworkflowrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/executeactionworkflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generatestate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

// createActionWorkflowStep creates a workflow step for executing an action workflow
func createActionWorkflowStep(ctx workflow.Context, installID string, iaw *app.InstallActionWorkflow, triggeredByID string, runEnvVars map[string]string, role string, sg *stepGroup) (*app.WorkflowStep, error) {
	sig := &executeactionworkflow.Signal{
		Signal: &actionworkflowrun.Signal{
			InstallID:               installID,
			InstallActionWorkflowID: iaw.ID,
			TriggerType:             app.ActionWorkflowTriggerTypeManual,
			TriggeredByID:           triggeredByID,
			TriggeredByType:         string(app.ActionWorkflowTriggerTypeManual),
			RunEnvVars:              runEnvVars,
			Role:                    role,
		},
	}

	name := fmt.Sprintf("%s action workflow run", string(app.ActionWorkflowTriggerTypeManual))
	return sg.installSignalStep(ctx, installID, name, pgtype.Hstore{}, sig, false)
}

func RunActionWorkflow(ctx workflow.Context, flw *app.Workflow) (*app.GenerateStepsResult, error) {
	sg := newStepGroup(flw)

	installID := flw.OwnerID
	if flw.OwnerType != "installs" {
		return nil, errors.New("invalid owner set on workflow")
	}

	steps := make([]*app.WorkflowStep, 0)

	adhocActionRunID, isAdhoc := flw.Metadata["adhoc_action_run_id"]
	if isAdhoc {
		return handleAdhocActionRun(ctx, flw, installID, adhocActionRunID, sg, steps)
	}

	installActionWorkflowID, ok := flw.Metadata["install_action_workflow_id"]
	if !ok {
		return nil, errors.New("install action workflow is not set on the install workflow for a manual deploy")
	}
	triggeredByID, ok := flw.Metadata["triggerred_by_id"]
	if !ok {
		return nil, errors.New("triggerred by id is not set on the install workflow for a manual deploy")
	}

	iaw, err := activities.AwaitGetInstallActionWorkflowByID(ctx, generics.FromPtrStr(installActionWorkflowID))
	if err != nil {
		return nil, err
	}

	prefix := "RUNENV_"
	runEnvVars := map[string]string{}

	for key, value := range flw.Metadata {
		if strings.HasPrefix(key, prefix) {
			newKey := key[len(prefix):]
			runEnvVars[newKey] = *value
		}
	}

	runEnvVars["TRIGGER_TYPE"] = string(app.ActionWorkflowTriggerTypeManual)

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	sg.nextGroupEager() // generate install state
	orgEnabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureStateGenV2))
	if err != nil {
		return nil, errors.Wrap(err, "unable to check state-gen-v2 feature")
	}
	stateGenV2 := statemanager.UseStateGenV2(orgEnabled, install.Metadata)
	if !stateGenV2 {
		stateSignal := &generatestate.Signal{InstallID: installID}
		step, err := sg.installSignalStep(ctx, installID, "generate install state", pgtype.Hstore{}, stateSignal, flw.PlanOnly, WithSkippable(false))
		if err != nil {
			return nil, err
		}
		steps = append(steps, step)
	}

	sg.nextGroup()
	step, err := createActionWorkflowStep(ctx, installID, iaw, generics.FromPtrStr(triggeredByID), runEnvVars, flw.Role, sg)
	if err != nil {
		return nil, err
	}

	steps = append(steps, step)
	return sg.Result(steps), nil
}

func handleAdhocActionRun(ctx workflow.Context, flw *app.Workflow, installID string, adhocActionRunID *string, sg *stepGroup, steps []*app.WorkflowStep) (*app.GenerateStepsResult, error) {
	actionName := generics.FromPtrStr(flw.Metadata["install_action_workflow_name"])
	if actionName == "" {
		actionName = "Adhoc action"
	}

	run, err := activities.AwaitGetInstallActionWorkflowRunByRunID(ctx, generics.FromPtrStr(adhocActionRunID))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get adhoc action run")
	}

	sig := &actionworkflowrun.Signal{
		InstallID:        installID,
		AdhocActionRunID: run.ID,
		Role:             flw.Role,
	}

	sg.nextGroup()
	stepName := fmt.Sprintf("Running adhoc action %s", actionName)
	step, err := sg.installSignalStep(ctx, installID, stepName, pgtype.Hstore{}, sig, flw.PlanOnly, WithExecutionType(app.WorkflowStepExecutionTypeUser))
	if err != nil {
		return nil, err
	}

	steps = append(steps, step)
	return sg.Result(steps), nil
}
