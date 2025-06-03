package flows

import (
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"go.temporal.io/sdk/workflow"
)

func installSignalStep(ctx workflow.Context, installID, name string, metadata pgtype.Hstore, signal *signals.Signal) (*app.FlowStep, error) {
	if signal == nil {
		return &app.FlowStep{
			Name:          name,
			ExecutionType: app.FlowStepExecutionTypeSkipped,
			Status:        app.NewCompositeTemporalStatus(ctx, app.StatusPending),
			Metadata:      metadata,
		}, nil
	}
	byts, err := json.Marshal(signal)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create signal json")
	}

	executionTyp := app.FlowStepExecutionTypeSystem
	// user signals
	userSignals := []eventloop.SignalType{
		signals.OperationAwaitInstallStackVersionRun,
	}
	if generics.SliceContains(signal.Type, userSignals) {
		executionTyp = app.FlowStepExecutionTypeUser
	}

	// await approval signals
	approvalSignals := []eventloop.SignalType{
		signals.OperationProvisionSandbox,
		signals.OperationDeprovisionSandbox,
		signals.OperationReprovisionSandbox,
		signals.OperationExecuteDeployComponent,
		signals.OperationExecuteTeardownComponent,
	}
	if generics.SliceContains(signal.Type, approvalSignals) {
		executionTyp = app.FlowStepExecutionTypeApproval
	}

	return &app.FlowStep{
		Name:          name,
		ExecutionType: executionTyp,
		Status:        app.NewCompositeTemporalStatus(ctx, app.StatusPending),
		Metadata:      metadata,
		Signal: app.Signal{
			Namespace:   "installs",
			Type:        string(signal.Type),
			EventLoopID: installID,
			SignalJSON:  byts,
		},
	}, nil
}

func getComponentLifecycleActionsSteps(ctx workflow.Context, flowID, componentID, installID string, triggerTyp app.ActionWorkflowTriggerType) ([]*app.FlowStep, error) {
	comp, err := activities.AwaitGetComponentByComponentID(ctx, componentID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component")
	}

	steps := make([]*app.FlowStep, 0)
	triggers, err := activities.AwaitGetInstallActionWorkflowsByTriggerType(ctx, activities.GetInstallActionWorkflowsByTriggerTypeRequest{
		ComponentID: comp.ID,
		InstallID:   installID,
		TriggerType: triggerTyp,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get components")
	}

	for _, trigger := range triggers {
		sig := &signals.Signal{
			Type: signals.OperationExecuteActionWorkflow,
			InstallActionWorkflowTrigger: signals.InstallActionWorkflowTriggerSubSignal{
				InstallActionWorkflowID: trigger.ID,
				TriggerType:             triggerTyp,
				TriggeredByID:           flowID,
				TriggeredByType:         string(triggerTyp),
				RunEnvVars: map[string]string{
					"TRIGGER_TYPE":   string(triggerTyp),
					"COMPONENT_ID":   comp.ID,
					"COMPONENT_NAME": comp.Name,
				},
			},
		}
		name := fmt.Sprintf("%s %s action workflow run", comp.Name, triggerTyp)
		step, err := installSignalStep(ctx, installID, name, pgtype.Hstore{}, sig)
		if err != nil {
			return nil, err
		}

		steps = append(steps, step)
	}

	return steps, nil
}

func getComponentDeploySteps(ctx workflow.Context, installID string, flw *app.Flow, componentIDs []string) ([]*app.FlowStep, error) {
	steps := make([]*app.FlowStep, 0)
	for _, compID := range componentIDs {
		comp, err := activities.AwaitGetComponentByComponentID(ctx, compID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get component")
		}

		preDeploySteps, err := getComponentLifecycleActionsSteps(ctx, flw.ID, compID, installID, app.ActionWorkflowTriggerTypePreDeployComponent)
		if err != nil {
			return nil, err
		}
		steps = append(steps, preDeploySteps...)

		deployStep, err := installSignalStep(ctx, installID, "deploy "+comp.Name, pgtype.Hstore{}, &signals.Signal{
			Type: signals.OperationExecuteDeployComponent,
			ExecuteDeployComponentSubSignal: signals.DeployComponentSubSignal{
				ComponentID: compID,
			},
		})
		steps = append(steps, deployStep)

		postDeploySteps, err := getComponentLifecycleActionsSteps(ctx, flw.ID, compID, installID, app.ActionWorkflowTriggerTypePostDeployComponent)
		if err != nil {
			return nil, err
		}
		steps = append(steps, postDeploySteps...)
	}

	return steps, nil
}

func getLifecycleActionsSteps(ctx workflow.Context, installID string, flw *app.Flow, triggerTyp app.ActionWorkflowTriggerType) ([]*app.FlowStep, error) {
	steps := make([]*app.FlowStep, 0)
	triggers, err := activities.AwaitGetInstallActionWorkflowsByTriggerType(ctx, activities.GetInstallActionWorkflowsByTriggerTypeRequest{
		InstallID:   installID,
		TriggerType: triggerTyp,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get components")
	}

	for _, trigger := range triggers {
		sig := &signals.Signal{
			Type: signals.OperationExecuteActionWorkflow,
			InstallActionWorkflowTrigger: signals.InstallActionWorkflowTriggerSubSignal{
				InstallActionWorkflowID: trigger.ID,
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
		name := fmt.Sprintf("%s action workflow run", triggerTyp)
		step, err := installSignalStep(ctx, installID, name, pgtype.Hstore{}, sig)
		if err != nil {
			return nil, err
		}

		steps = append(steps, step)
	}

	return steps, nil
}

func deployAllComponents(ctx workflow.Context, installID string, flw *app.Flow) ([]*app.FlowStep, error) {
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

	deploySteps, err := getComponentDeploySteps(ctx, installID, flw, componentIDs)
	if err != nil {
		return nil, err
	}

	steps = append(steps, deploySteps...)
	return steps, nil
}
