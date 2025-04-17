package worker

import (
	"encoding/json"
	"fmt"
	"strconv"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

func GenerateWorkflowStepsWorkflowID(req signals.RequestSignal) string {
	return fmt.Sprintf("%s-generate-workflow-steps", req.WorkflowID(req.ID))
}

// @id-callback GenerateWorkflowStepsWorkflowID
// @temporal-gen workflow
// @execution-timeout 120m
func (w *Workflows) GenerateWorkflowSteps(ctx workflow.Context, sreq signals.RequestSignal) error {
	workflowID := sreq.InstallWorkflowID

	wkflow, err := activities.AwaitGetInstallWorkflowByID(ctx, workflowID)
	if err != nil {
		return errors.Wrap(err, "unable to get install workflow")
	}

	steps, err := w.getSteps(ctx, wkflow)
	if err != nil {
		return errors.Wrap(err, "unable to generate steps")
	}
	for idx, step := range steps {
		if err := activities.AwaitCreateInstallWorkflowStep(ctx, activities.CreateInstallWorkflowStepRequest{
			InstallWorkflowID: sreq.InstallWorkflowID,
			InstallID:         step.InstallID,
			Status:            step.Status,
			Name:              step.Name,
			Signal:            step.Signal,
			Idx:               idx,
		}); err != nil {
			return errors.Wrap(err, "unable to create steps")
		}
	}

	return nil
}

func (w *Workflows) getSteps(ctx workflow.Context, wkflow *app.InstallWorkflow) ([]*app.InstallWorkflowStep, error) {
	switch wkflow.Type {
	case app.InstallWorkflowTypeManualDeploy:
		return w.getManualDeploySteps(ctx, wkflow)
	case app.InstallWorkflowTypeDeployComponents:
		return w.deployAllComponents(ctx, wkflow)
	case app.InstallWorkflowTypeTeardownComponents:
		return w.getComponentTeardownSteps(ctx, wkflow)
	case app.InstallWorkflowTypeInputUpdate:
		return w.getUpdateInputSteps(ctx, wkflow)
		// overall lifecycle
	case app.InstallWorkflowTypeProvision:
		return w.getInstallWorkflowProvisionSteps(ctx, wkflow)
	case app.InstallWorkflowTypeReprovision:
		return w.getInstallWorkflowReprovisionSteps(ctx, wkflow)
	case app.InstallWorkflowTypeReprovisionSandbox:
		return w.getInstallWorkflowReprovisionSandboxSteps(ctx, wkflow)
	case app.InstallWorkflowTypeDeprovision:
		return w.getInstallWorkflowDeprovisionSteps(ctx, wkflow)
	}

	return nil, nil
}

func (w *Workflows) getInstallWorkflowProvisionSteps(ctx workflow.Context, wkflow *app.InstallWorkflow) ([]*app.InstallWorkflowStep, error) {
	steps := make([]*app.InstallWorkflowStep, 0)

	step, err := w.installSignalStep(ctx, wkflow.InstallID, "provision runner", &signals.Signal{
		Type: signals.OperationProvisionRunner,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = w.installSignalStep(ctx, wkflow.InstallID, "generate install stack", &signals.Signal{
		Type: signals.OperationGenerateInstallStackVersion,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = w.installSignalStep(ctx, wkflow.InstallID, "await install stack", &signals.Signal{
		Type: signals.OperationAwaitInstallStackVersionRun,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = w.installSignalStep(ctx, wkflow.InstallID, "update install stack outputs", &signals.Signal{
		Type: signals.OperationUpdateInstallStackOutputs,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = w.installSignalStep(ctx, wkflow.InstallID, "await runner health", &signals.Signal{
		Type: signals.OperationAwaitRunnerHealthy,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err := w.getSandboxLifecycleActionsSteps(ctx, wkflow.ID, wkflow.InstallID, app.ActionWorkflowTriggerTypePreSandboxRun)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	step, err = w.installSignalStep(ctx, wkflow.InstallID, "provision sandbox", &signals.Signal{
		Type: signals.OperationProvisionSandbox,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err = w.getSandboxLifecycleActionsSteps(ctx, wkflow.ID, wkflow.InstallID, app.ActionWorkflowTriggerTypePostSandboxRun)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	deploySteps, err := w.deployAllComponents(ctx, wkflow)
	if err != nil {
		return nil, err
	}
	steps = append(steps, deploySteps...)

	return steps, nil
}

func (w *Workflows) getInstallWorkflowReprovisionSteps(ctx workflow.Context, wkflow *app.InstallWorkflow) ([]*app.InstallWorkflowStep, error) {
	steps := make([]*app.InstallWorkflowStep, 0)

	step, err := w.installSignalStep(ctx, wkflow.InstallID, "reprovision runner", &signals.Signal{
		Type: signals.OperationReprovisionRunner,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = w.installSignalStep(ctx, wkflow.InstallID, "generate install stack", &signals.Signal{
		Type: signals.OperationGenerateInstallStackVersion,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = w.installSignalStep(ctx, wkflow.InstallID, "await install stack", &signals.Signal{
		Type: signals.OperationAwaitInstallStackVersionRun,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = w.installSignalStep(ctx, wkflow.InstallID, "update install stack outputs", &signals.Signal{
		Type: signals.OperationUpdateInstallStackOutputs,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	step, err = w.installSignalStep(ctx, wkflow.InstallID, "await runner health", &signals.Signal{
		Type: signals.OperationAwaitRunnerHealthy,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err := w.getSandboxLifecycleActionsSteps(ctx, wkflow.ID, wkflow.InstallID, app.ActionWorkflowTriggerTypePreSandboxRun)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	step, err = w.installSignalStep(ctx, wkflow.InstallID, "reprovision sandbox", &signals.Signal{
		Type: signals.OperationReprovisionSandbox,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err = w.getSandboxLifecycleActionsSteps(ctx, wkflow.ID, wkflow.InstallID, app.ActionWorkflowTriggerTypePostSandboxRun)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	deploySteps, err := w.deployAllComponents(ctx, wkflow)
	if err != nil {
		return nil, err
	}
	steps = append(steps, deploySteps...)

	return steps, nil
}

func (w *Workflows) getInstallWorkflowDeleteSteps(ctx workflow.Context, wkflow *app.InstallWorkflow) ([]*app.InstallWorkflowStep, error) {
	steps, err := w.getInstallWorkflowDeprovisionSteps(ctx, wkflow)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get workflow deprovision steps")
	}

	step, err := w.installSignalStep(ctx, wkflow.InstallID, "delete install", &signals.Signal{
		Type: signals.OperationDelete,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	return steps, nil
}

func (w *Workflows) getInstallWorkflowDeprovisionSteps(ctx workflow.Context, wkflow *app.InstallWorkflow) ([]*app.InstallWorkflowStep, error) {
	steps := make([]*app.InstallWorkflowStep, 0)

	step, err := w.installSignalStep(ctx, wkflow.InstallID, "await runner healthy", &signals.Signal{
		Type: signals.OperationAwaitRunnerHealthy,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	deploySteps, err := w.getComponentTeardownSteps(ctx, wkflow)
	if err != nil {
		return nil, err
	}
	steps = append(steps, deploySteps...)

	step, err = w.installSignalStep(ctx, wkflow.InstallID, "deprovision sandbox", &signals.Signal{
		Type: signals.OperationDeprovisionSandbox,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	return steps, nil
}

func (w *Workflows) getComponentLifecycleActionsSteps(ctx workflow.Context, installWorkflowID, componentID, installID string, triggerTyp app.ActionWorkflowTriggerType) ([]*app.InstallWorkflowStep, error) {
	comp, err := activities.AwaitGetComponentByComponentID(ctx, componentID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component")
	}

	steps := make([]*app.InstallWorkflowStep, 0)
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
				TriggeredByID:           installWorkflowID,
				RunEnvVars: map[string]string{
					"TRIGGER_TYPE":   string(triggerTyp),
					"COMPONENT_ID":   comp.ID,
					"COMPONENT_NAME": comp.Name,
				},
			},
		}
		name := fmt.Sprintf("%s %s action workflow run", comp.Name, triggerTyp)
		step, err := w.installSignalStep(ctx, installID, name, sig)
		if err != nil {
			return nil, err
		}

		steps = append(steps, step)
	}

	return steps, nil
}

func (w *Workflows) getComponentDeploySteps(ctx workflow.Context, wkflow *app.InstallWorkflow, componentIDs []string) ([]*app.InstallWorkflowStep, error) {
	steps := make([]*app.InstallWorkflowStep, 0)
	for _, compID := range componentIDs {
		comp, err := activities.AwaitGetComponentByComponentID(ctx, compID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get component")
		}

		preDeploySteps, err := w.getComponentLifecycleActionsSteps(ctx, wkflow.ID, compID, wkflow.InstallID, app.ActionWorkflowTriggerTypePreDeployComponent)
		if err != nil {
			return nil, err
		}
		steps = append(steps, preDeploySteps...)

		deployStep, err := w.installSignalStep(ctx, wkflow.InstallID, "deploy "+comp.Name, &signals.Signal{
			Type: signals.OperationExecuteDeployComponent,
			ExecuteDeployComponentSubSignal: signals.DeployComponentSubSignal{
				ComponentID: compID,
			},
		})
		steps = append(steps, deployStep)

		postDeploySteps, err := w.getComponentLifecycleActionsSteps(ctx, wkflow.ID, compID, wkflow.InstallID, app.ActionWorkflowTriggerTypePostDeployComponent)
		if err != nil {
			return nil, err
		}
		steps = append(steps, postDeploySteps...)
	}

	return steps, nil
}

func (w *Workflows) getManualDeploySteps(ctx workflow.Context, wkflow *app.InstallWorkflow) ([]*app.InstallWorkflowStep, error) {
	steps := make([]*app.InstallWorkflowStep, 0)
	step, err := w.installSignalStep(ctx, wkflow.InstallID, "await runner healthy", &signals.Signal{
		Type: signals.OperationAwaitRunnerHealthy,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	installDeployID, ok := wkflow.Metadata["install_deploy_id"]
	if !ok {
		return nil, errors.New("install deploy is not set on the install workflow for a manual deploy")
	}

	deployDependents, _ := wkflow.Metadata["deploy_dependents"]

	installDeploy, err := activities.AwaitGetDeployByDeployID(ctx, generics.FromPtrStr(installDeployID))
	if err != nil {
		return nil, errors.New("unable to get install deploy")
	}
	install, err := activities.AwaitGetByInstallID(ctx, wkflow.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	// first, provision the deploy with before and after triggers
	comp, err := activities.AwaitGetComponentByComponentID(ctx, installDeploy.ComponentID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component")
	}

	preDeploySteps, err := w.getComponentLifecycleActionsSteps(ctx, wkflow.ID, installDeploy.ComponentID, wkflow.InstallID, app.ActionWorkflowTriggerTypePreDeployComponent)
	if err != nil {
		return nil, err
	}
	steps = append(steps, preDeploySteps...)

	deployStep, err := w.installSignalStep(ctx, wkflow.InstallID, "deploy "+comp.Name, &signals.Signal{
		Type: signals.OperationExecuteDeployComponent,
		ExecuteDeployComponentSubSignal: signals.DeployComponentSubSignal{
			DeployID:    generics.FromPtrStr(installDeployID),
			ComponentID: comp.ID,
		},
	})
	steps = append(steps, deployStep)

	postDeploySteps, err := w.getComponentLifecycleActionsSteps(ctx, wkflow.ID, installDeploy.ComponentID, wkflow.InstallID, app.ActionWorkflowTriggerTypePostDeployComponent)
	if err != nil {
		return nil, err
	}
	steps = append(steps, postDeploySteps...)

	// now queue up any deploy that _depend_ on the input
	componentIDs, err := activities.AwaitGetAppInstallGraph(ctx, activities.GetAppInstallGraphRequest{
		AppID:     install.AppID,
		InstallID: install.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install graph")
	}

	dependencyCompIDs := generics.SliceAfterValue(componentIDs, comp.ID)
	dependencyDeploySteps, err := w.getComponentDeploySteps(ctx, wkflow, dependencyCompIDs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component deploy steps")
	}

	if generics.FromPtrStr(deployDependents) == strconv.FormatBool(true) {
		steps = append(steps, dependencyDeploySteps...)
	}

	return nil, nil
}

func (w *Workflows) getComponentTeardownSteps(ctx workflow.Context, wkflow *app.InstallWorkflow) ([]*app.InstallWorkflowStep, error) {
	install, err := activities.AwaitGetByInstallID(ctx, wkflow.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	componentIDs, err := activities.AwaitGetAppInstallGraph(ctx, activities.GetAppInstallGraphRequest{
		AppID:     install.AppID,
		InstallID: install.ID,
		Reverse:   true,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install graph")
	}

	steps := make([]*app.InstallWorkflowStep, 0)
	for _, compID := range componentIDs {
		comp, err := activities.AwaitGetComponentByComponentID(ctx, compID)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get component")
		}

		preDeploySteps, err := w.getComponentLifecycleActionsSteps(ctx, wkflow.ID, compID, wkflow.InstallID, app.ActionWorkflowTriggerTypePreTeardownComponent)
		if err != nil {
			return nil, err
		}
		steps = append(steps, preDeploySteps...)

		deployStep, err := w.installSignalStep(ctx, wkflow.InstallID, "teardown "+comp.Name, &signals.Signal{
			Type: signals.OperationExecuteTeardownComponent,
			ExecuteDeployComponentSubSignal: signals.DeployComponentSubSignal{
				ComponentID: compID,
			},
		})
		steps = append(steps, deployStep)

		postDeploySteps, err := w.getComponentLifecycleActionsSteps(ctx, wkflow.ID, compID, wkflow.InstallID, app.ActionWorkflowTriggerTypePostTeardownComponent)
		if err != nil {
			return nil, err
		}
		steps = append(steps, postDeploySteps...)
	}

	return steps, nil
}

func (w *Workflows) installSignalStep(ctx workflow.Context, installID, name string, signal *signals.Signal) (*app.InstallWorkflowStep, error) {
	byts, err := json.Marshal(signal)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create signal json")
	}

	return &app.InstallWorkflowStep{
		Name:      name,
		InstallID: installID,
		Status:    app.NewCompositeTemporalStatus(ctx, app.StatusPending),
		Signal: app.Signal{
			Namespace:   "installs",
			Type:        string(signal.Type),
			EventLoopID: installID,
			SignalJSON:  byts,
		},
	}, nil
}

// update input steps will _eventually_ update only the components that depend on it, but for now we just trigger _all_
// components, as they can be denied by the user.
func (w *Workflows) getUpdateInputSteps(ctx workflow.Context, wkflow *app.InstallWorkflow) ([]*app.InstallWorkflowStep, error) {
	install, err := activities.AwaitGetByInstallID(ctx, wkflow.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	componentIDs, err := activities.AwaitGetAppInstallGraph(ctx, activities.GetAppInstallGraphRequest{
		AppID:     install.AppID,
		InstallID: install.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install graph")
	}

	return w.getComponentDeploySteps(ctx, wkflow, componentIDs)
}

func (w *Workflows) deployAllComponents(ctx workflow.Context, wkflow *app.InstallWorkflow) ([]*app.InstallWorkflowStep, error) {
	install, err := activities.AwaitGetByInstallID(ctx, wkflow.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	componentIDs, err := activities.AwaitGetAppInstallGraph(ctx, activities.GetAppInstallGraphRequest{
		AppID:     install.AppID,
		InstallID: install.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install graph")
	}

	steps := make([]*app.InstallWorkflowStep, 0)
	step, err := w.installSignalStep(ctx, wkflow.InstallID, "await runner healthy", &signals.Signal{
		Type: signals.OperationAwaitRunnerHealthy,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	deploySteps, err := w.getComponentDeploySteps(ctx, wkflow, componentIDs)
	if err != nil {
		return nil, err
	}

	steps = append(steps, deploySteps...)
	return steps, nil
}

func (w *Workflows) getSandboxLifecycleActionsSteps(ctx workflow.Context, installWorkflowID, installID string, triggerTyp app.ActionWorkflowTriggerType) ([]*app.InstallWorkflowStep, error) {
	steps := make([]*app.InstallWorkflowStep, 0)
	triggers, err := activities.AwaitGetInstallActionWorkflowsByTriggerType(ctx, activities.GetInstallActionWorkflowsByTriggerTypeRequest{
		InstallID:   installID,
		TriggerType: triggerTyp,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get components")
	}

	workflow, err := activities.AwaitGetInstallWorkflowByID(ctx, installWorkflowID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install workflow")
	}

	for _, trigger := range triggers {
		sig := &signals.Signal{
			Type: signals.OperationExecuteActionWorkflow,
			InstallActionWorkflowTrigger: signals.InstallActionWorkflowTriggerSubSignal{
				InstallActionWorkflowID: trigger.ID,
				TriggerType:             triggerTyp,
				TriggeredByID:           installWorkflowID,
				RunEnvVars: map[string]string{
					"TRIGGER_TYPE":          string(triggerTyp),
					"INSTALL_WORKFLOW_TYPE": string(workflow.Type),
					"INSTALL_WORKFLOW_ID":   workflow.ID,
				},
			},
		}
		name := fmt.Sprintf("%s action workflow run", triggerTyp)
		step, err := w.installSignalStep(ctx, installID, name, sig)
		if err != nil {
			return nil, err
		}

		steps = append(steps, step)
	}

	return steps, nil
}

func (w *Workflows) teardownAllComponents(ctx workflow.Context, wkflow *app.InstallWorkflow) ([]*app.InstallWorkflowStep, error) {
	steps := make([]*app.InstallWorkflowStep, 0)
	step, err := w.installSignalStep(ctx, wkflow.InstallID, "await runner healthy", &signals.Signal{
		Type: signals.OperationAwaitRunnerHealthy,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	deploySteps, err := w.getComponentTeardownSteps(ctx, wkflow)
	if err != nil {
		return nil, err
	}

	steps = append(steps, deploySteps...)
	return steps, nil
}

func (w *Workflows) getInstallWorkflowReprovisionSandboxSteps(ctx workflow.Context, wkflow *app.InstallWorkflow) ([]*app.InstallWorkflowStep, error) {
	steps := make([]*app.InstallWorkflowStep, 0)

	step, err := w.installSignalStep(ctx, wkflow.InstallID, "await runner health", &signals.Signal{
		Type: signals.OperationAwaitRunnerHealthy,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err := w.getSandboxLifecycleActionsSteps(ctx, wkflow.ID, wkflow.InstallID, app.ActionWorkflowTriggerTypePreSandboxRun)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	step, err = w.installSignalStep(ctx, wkflow.InstallID, "reprovision sandbox", &signals.Signal{
		Type: signals.OperationReprovisionSandbox,
	})
	if err != nil {
		return nil, err
	}
	steps = append(steps, step)

	lifecycleSteps, err = w.getSandboxLifecycleActionsSteps(ctx, wkflow.ID, wkflow.InstallID, app.ActionWorkflowTriggerTypePostSandboxRun)
	if err != nil {
		return nil, err
	}
	steps = append(steps, lifecycleSteps...)

	deploySteps, err := w.deployAllComponents(ctx, wkflow)
	if err != nil {
		return nil, err
	}
	steps = append(steps, deploySteps...)

	return steps, nil
}
