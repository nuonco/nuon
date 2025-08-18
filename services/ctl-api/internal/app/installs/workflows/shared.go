package workflows

import (
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

func installSignalStep(ctx workflow.Context, installID, name string, metadata pgtype.Hstore, signal *signals.Signal, planOnly bool, opts ...WorkflowStepOptions) (*app.WorkflowStep, error) {
	if signal == nil {
		return &app.WorkflowStep{
			Name:          name,
			ExecutionType: app.WorkflowStepExecutionTypeSkipped,
			Status:        app.NewCompositeTemporalStatus(ctx, app.StatusPending),
			Metadata:      metadata,
		}, nil
	}
	byts, err := json.Marshal(signal)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create signal json")
	}

	var targettype string
	retryable := true

	switch signal.Type {
	case signals.OperationAwaitInstallStackVersionRun, signals.OperationGenerateInstallStackVersion, signals.OperationUpdateInstallStackOutputs:
		targettype = "install_stack_versions"
		retryable = false
	case signals.OperationAwaitRunnerHealthy:
		targettype = "runners"
		retryable = false
	case signals.OperationExecuteDeployComponentApplyPlan,
		signals.OperationExecuteDeployComponentSyncAndPlan,
		signals.OperationExecuteDeployComponentSyncImage,
		signals.OperationExecuteTeardownComponentSyncAndPlan,
		signals.OperationExecuteTeardownComponentApplyPlan:
		targettype = "install_deploys"
	case signals.OperationProvisionSandboxPlan,
		signals.OperationProvisionSandboxApplyPlan,
		signals.OperationDeprovisionSandboxPlan,
		signals.OperationDeprovisionSandboxApplyPlan,
		signals.OperationReprovisionSandboxPlan,
		signals.OperationReprovisionSandboxApplyPlan:
		targettype = "install_sandbox_runs"
	case signals.OperationExecuteActionWorkflow:
		targettype = "install_action_workflow_runs"
	case signals.OperationGenerateState:
		targettype = "install_states"
	}

	executionTyp := app.WorkflowStepExecutionTypeSystem
	// user signals

	userSignals := []eventloop.SignalType{
		signals.OperationAwaitInstallStackVersionRun,
	}
	if generics.SliceContains(signal.Type, userSignals) {
		executionTyp = app.WorkflowStepExecutionTypeUser
	}

	// await approval signals
	approvalSignals := []eventloop.SignalType{
		signals.OperationProvisionSandboxPlan,
		signals.OperationDeprovisionSandboxPlan,
		signals.OperationReprovisionSandboxPlan,
		signals.OperationExecuteDeployComponentSyncAndPlan,
		signals.OperationExecuteTeardownComponentSyncAndPlan,
	}
	if generics.SliceContains(signal.Type, approvalSignals) {
		executionTyp = app.WorkflowStepExecutionTypeApproval
	}

	// plan-only-skip signals are signals that should not be executed, when in plan only
	planOnlySkipSignals := []eventloop.SignalType{
		signals.OperationDeprovisionSandboxApplyPlan,
		signals.OperationProvisionSandboxApplyPlan,
		signals.OperationReprovisionSandboxApplyPlan,
		signals.OperationExecuteDeployComponentApplyPlan,
		signals.OperationExecuteTeardownComponentApplyPlan,
	}
	if planOnly && generics.SliceContains(signal.Type, planOnlySkipSignals) {
		executionTyp = app.WorkflowStepExecutionTypeSkipped
	}

	if signal.Type == signals.OperationGenerateState {
		executionTyp = app.WorkflowStepExecutionTypeHidden
	}

	s := &app.WorkflowStep{
		Name:           name,
		ExecutionType:  executionTyp,
		StepTargetType: targettype,
		OwnerID:        installID,
		OwnerType:      "installs",
		Status:         app.NewCompositeTemporalStatus(ctx, app.StatusPending),
		Metadata:       metadata,
		Signal: app.Signal{
			Namespace:   "installs",
			Type:        string(signal.Type),
			EventLoopID: installID,
			SignalJSON:  byts,
		},
		Retryable: retryable,
		Skippable: true,
	}

	for _, o := range opts {
		o(s)
	}

	return s, nil
}

func getComponentLifecycleActionsSteps(ctx workflow.Context, flw *app.Workflow, comp *app.Component, installID string, triggerTyp app.ActionWorkflowTriggerType, sg *stepGroup) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0)
	installActions, err := activities.AwaitGetInstallActionWorkflowsByTriggerType(ctx, activities.GetInstallActionWorkflowsByTriggerTypeRequest{
		ComponentID: comp.ID,
		InstallID:   installID,
		TriggerType: triggerTyp,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get action workflows")
	}

	sg.nextGroup() // lifecycleSteps

	for _, installAction := range installActions {
		sig := &signals.Signal{
			Type: signals.OperationExecuteActionWorkflow,
			InstallActionWorkflowTrigger: signals.InstallActionWorkflowTriggerSubSignal{
				InstallActionWorkflowID: installAction.ID,
				TriggerType:             triggerTyp,
				TriggeredByID:           flw.ID,
				TriggeredByType:         string(triggerTyp),
				RunEnvVars: map[string]string{
					"TRIGGER_TYPE":   string(triggerTyp),
					"COMPONENT_ID":   comp.ID,
					"COMPONENT_NAME": comp.Name,
				},
			},
		}
		name := fmt.Sprintf("%s Action Run (%s)", installAction.ActionWorkflow.Name, triggerTyp)
		step, err := sg.installSignalStep(ctx, installID, name, pgtype.Hstore{}, sig, flw.PlanOnly)
		if err != nil {
			return nil, err
		}

		steps = append(steps, step)
	}

	return steps, nil
}

func getComponentDeploySteps(ctx workflow.Context, installID string, flw *app.Workflow, componentIDs []string, sg *stepGroup) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0)

	install, err := activities.AwaitGetByInstallID(ctx, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	appcfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}
	components := make(map[string]app.Component)
	for _, ccc := range appcfg.ComponentConfigConnections {
		components[ccc.ComponentID] = ccc.Component
	}

	for _, compID := range componentIDs {
		comp, has := components[compID]
		if !has {
			return nil, errors.Errorf("component %s not found in app config", compID)
		}

		preDeploySteps, err := getComponentLifecycleActionsSteps(ctx, flw, &comp, installID, app.ActionWorkflowTriggerTypePreDeployComponent, sg)
		if err != nil {
			return nil, err
		}
		steps = append(steps, preDeploySteps...)

		// sync image
		if comp.Type.IsImage() {
			sg.nextGroup() // sync
			deployStep, err := sg.installSignalStep(ctx, installID, "sync "+comp.Name, pgtype.Hstore{}, &signals.Signal{
				Type: signals.OperationExecuteDeployComponentSyncImage,
				ExecuteDeployComponentSubSignal: signals.DeployComponentSubSignal{
					ComponentID: comp.ID,
				},
			}, flw.PlanOnly)
			if err != nil {
				return nil, errors.Wrap(err, "unable to create image sync")
			}

			steps = append(steps, deployStep)
		} else {
			sg.nextGroup() // component sync + plan + apply
			planStep, err := sg.installSignalStep(ctx, installID, "sync and plan "+comp.Name, pgtype.Hstore{}, &signals.Signal{
				Type: signals.OperationExecuteDeployComponentSyncAndPlan,
				ExecuteDeployComponentSubSignal: signals.DeployComponentSubSignal{
					ComponentID: comp.ID,
				},
			}, flw.PlanOnly, WithSkippable(false))
			if err != nil {
				return nil, errors.Wrap(err, "unable to create image sync")
			}

			applyPlanStep, err := sg.installSignalStep(ctx, installID, "apply "+comp.Name, pgtype.Hstore{}, &signals.Signal{
				Type: signals.OperationExecuteDeployComponentApplyPlan,
				ExecuteDeployComponentSubSignal: signals.DeployComponentSubSignal{
					ComponentID: comp.ID,
				},
			}, flw.PlanOnly)
			if err != nil {
				return nil, errors.Wrap(err, "unable to create image sync")
			}
			steps = append(steps, planStep, applyPlanStep)
		}
		postDeploySteps, err := getComponentLifecycleActionsSteps(ctx, flw, &comp, installID, app.ActionWorkflowTriggerTypePostDeployComponent, sg)
		if err != nil {
			return nil, err
		}
		steps = append(steps, postDeploySteps...)
	}

	return steps, nil
}

func getLifecycleActionsSteps(ctx workflow.Context, installID string, flw *app.Workflow, triggerTyp app.ActionWorkflowTriggerType, sg *stepGroup) ([]*app.WorkflowStep, error) {
	steps := make([]*app.WorkflowStep, 0)

	installActions, err := activities.AwaitGetInstallActionWorkflowsByTriggerType(ctx, activities.GetInstallActionWorkflowsByTriggerTypeRequest{
		InstallID:   installID,
		TriggerType: triggerTyp,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get action workflows")
	}

	sg.nextGroup() // lifecycleSteps

	for _, installAction := range installActions {
		sig := &signals.Signal{
			Type: signals.OperationExecuteActionWorkflow,
			InstallActionWorkflowTrigger: signals.InstallActionWorkflowTriggerSubSignal{
				InstallActionWorkflowID: installAction.ID,
				TriggerType:             triggerTyp,
				TriggeredByID:           flw.ID,
				TriggeredByType:         string(triggerTyp),
				RunEnvVars: map[string]string{
					"TRIGGER_TYPE": string(triggerTyp),
					"FLOW_TYPE":    string(flw.Type),
					"FLOW_ID":      flw.ID,
					// TODO(sdboyer) remove these once they're updated on the other end
					"INSTALL_WORKFLOW_TYPE": string(flw.Type),
					"INSTALL_WORKFLOW_ID":   flw.ID,
				},
			},
		}
		name := fmt.Sprintf("%s Action Run (%s)", installAction.ActionWorkflow.Name, triggerTyp)
		step, err := sg.installSignalStep(ctx, installID, name, pgtype.Hstore{}, sig, flw.PlanOnly)
		if err != nil {
			return nil, err
		}

		steps = append(steps, step)
	}

	return steps, nil
}

func deployAllComponents(ctx workflow.Context, installID string, flw *app.Workflow, sg *stepGroup) ([]*app.WorkflowStep, error) {
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

	steps := make([]*app.WorkflowStep, 0)

	sg.nextGroup() // runner health

	step, err := sg.installSignalStep(ctx, installID, "await runner healthy", pgtype.Hstore{}, &signals.Signal{
		Type: signals.OperationAwaitRunnerHealthy,
	}, flw.PlanOnly)
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err := getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePreDeployAllComponents, sg)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)
	deploySteps, err := getComponentDeploySteps(ctx, installID, flw, componentIDs, sg)
	if err != nil {
		return nil, err
	}
	steps = append(steps, deploySteps...)
	lifecycleSteps, err = getLifecycleActionsSteps(ctx, installID, flw, app.ActionWorkflowTriggerTypePostDeployAllComponents, sg)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	return steps, nil
}
