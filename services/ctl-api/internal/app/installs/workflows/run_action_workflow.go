package workflows

import (
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

// createActionWorkflowStep creates a workflow step for executing an action workflow
func createActionWorkflowStep(ctx workflow.Context, installID string, iaw *app.InstallActionWorkflow, triggeredByID string, runEnvVars map[string]string, role string, sg *stepGroup) (*app.WorkflowStep, error) {
	sig := &signals.Signal{
		Type: signals.OperationExecuteActionWorkflow,
		InstallActionWorkflowTrigger: signals.InstallActionWorkflowTriggerSubSignal{
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

func RunActionWorkflow(ctx workflow.Context, flw *app.Workflow) ([]*app.WorkflowStep, error) {
	installID := generics.FromPtrStr(flw.Metadata["install_id"])
	sg := newStepGroup()

	steps := make([]*app.WorkflowStep, 0)
	sg.nextGroup()
	step, err := sg.installSignalStep(ctx, installID, "generate install state", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationGenerateState,
	}, flw.PlanOnly, WithSkippable(false))
	if err != nil {
		return nil, err
	}

	steps = append(steps, step)

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

	sg.nextGroup()
	step, err = createActionWorkflowStep(ctx, installID, iaw, generics.FromPtrStr(triggeredByID), runEnvVars, flw.Role, sg)
	if err != nil {
		return nil, err
	}

	steps = append(steps, step)
	return steps, nil
}

func handleAdhocActionRun(ctx workflow.Context, flw *app.Workflow, installID string, adhocActionRunID *string, sg *stepGroup, steps []*app.WorkflowStep) ([]*app.WorkflowStep, error) {
	actionName := generics.FromPtrStr(flw.Metadata["install_action_workflow_name"])
	if actionName == "" {
		actionName = "Adhoc action"
	}

	run, err := activities.AwaitGetInstallActionWorkflowRunByRunID(ctx, generics.FromPtrStr(adhocActionRunID))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get adhoc action run")
	}

	sig := &signals.Signal{
		Type:                signals.OperationActionWorkflowRun,
		ActionWorkflowRunID: run.ID,
	}

	sg.nextGroup()
	stepName := fmt.Sprintf("Running adhoc action %s", actionName)
	step, err := sg.installSignalStep(ctx, installID, stepName, pgtype.Hstore{}, sig, flw.PlanOnly, WithExecutionType(app.WorkflowStepExecutionTypeUser))
	if err != nil {
		return nil, err
	}

	steps = append(steps, step)
	return steps, nil
}
