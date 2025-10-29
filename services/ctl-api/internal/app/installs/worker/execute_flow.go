package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/workflows"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/flow"
)

func ExecuteWorkflowIDCallback(req signals.RequestSignal) string {
	return fmt.Sprintf("%s-execute-workflow-%s", req.ID, req.InstallWorkflowID)
}

// @temporal-gen workflow
// @execution-timeout 720h
// @id-callback ExecuteWorkflowIDCallback
func (w *Workflows) ExecuteFlow(ctx workflow.Context, sreq signals.RequestSignal) error {
	if sreq.FlowID == "" {
		sreq.FlowID = sreq.InstallWorkflowID
	}
	fc := &flow.WorkflowConductor[*signals.Signal]{
		Cfg:        w.cfg,
		V:          w.v,
		MW:         w.mw,
		Generators: w.getWorkflowStepGenerators(ctx),
		ExecFn:     w.getExecuteFlowExecFn(sreq),
	}

	err := fc.Handle(ctx, sreq.EventLoopRequest, sreq.FlowID, sreq.ID, sreq.StartFromStepIdx)
	if err != nil {
		cerr, ok := err.(*flow.ContinueAsNewErr)
		if ok && cerr != nil {
			sreq.StartFromStepIdx = cerr.StartFromStepIdx
			return workflow.NewContinueAsNewError(ctx, w.ExecuteFlow, sreq)
		}
		return err
	}
	return nil
}

func (w *Workflows) getWorkflowStepGenerators(ctx workflow.Context) map[app.WorkflowType]flow.WorkflowStepGenerator {
	return map[app.WorkflowType]flow.WorkflowStepGenerator{
		app.WorkflowTypeManualDeploy:               workflows.ManualDeploySteps,
		app.WorkflowTypeDriftRun:                   workflows.ManualDeploySteps,
		app.WorkflowTypeDeployComponents:           workflows.DeployAllComponents,
		app.WorkflowTypeTeardownComponent:          workflows.TeardownComponent,
		app.WorkflowTypeTeardownComponents:         workflows.TeardownComponents,
		app.WorkflowTypeInputUpdate:                workflows.InputUpdate,
		app.WorkflowTypeActionWorkflowRun:          workflows.RunActionWorkflow,
		app.WorkflowTypeProvision:                  workflows.Provision,
		app.WorkflowTypeReprovision:                workflows.Reprovision,
		app.WorkflowTypeReprovisionSandbox:         workflows.ReprovisionSandbox,
		app.WorkflowTypeDriftRunReprovisionSandbox: workflows.ReprovisionSandbox,
		app.WorkflowTypeDeprovision:                workflows.Deprovision,
		app.WorkflowTypeDeprovisionSandbox:         workflows.DeprovisionSandbox,
		app.WorkflowTypeSyncSecrets:                workflows.SyncSecrets,
	}
}

func (w *Workflows) getExecuteFlowExecFn(sreq signals.RequestSignal) func(workflow.Context, eventloop.EventLoopRequest, *signals.Signal, app.WorkflowStep) error {
	return func(ctx workflow.Context, ereq eventloop.EventLoopRequest, sig *signals.Signal, step app.WorkflowStep) error {
		sig.InstallWorkflowID = sreq.FlowID
		sig.FlowID = sreq.FlowID
		sig.WorkflowStepID = step.ID
		sig.WorkflowStepName = step.Name
		sig.FlowStepID = step.ID
		sig.FlowStepName = step.Name
		handlerSreq := signals.NewRequestSignal(ereq, sig)

		// Predict the workflow ID of the underlying object's event loop
		suffix, err := w.subloopSuffix(ctx, handlerSreq)
		if err != nil {
			return err
		}

		if suffix != "" {
			id := fmt.Sprintf("%s-%s", sreq.ID, suffix)
			if err := w.evClient.SendAndWait(ctx, id, &handlerSreq); err != nil {
				return err
			}
		} else {
			// no suffix means a workflow on this loop, so we must invoke the handler directly
			handlers := w.getHandlers()
			handler, ok := handlers[sig.Type]
			if !ok {
				return errors.New(fmt.Sprintf("no handler found for signal %s", sig.Type))
			}
			if err := handler(ctx, handlerSreq); err != nil {
				return err
			}
		}
		return nil
	}
}

// NOTE(sdboyer) this method is tightly coupled to the subloop naming logic in ./startup.go
func (w *Workflows) subloopSuffix(ctx workflow.Context, sreq signals.RequestSignal) (string, error) {
	// All errors _should_ be unreachable because these activities succeeded when bootstrapping the sub event loops
	if _, has := w.subwfStack.GetHandlers()[sreq.Type]; has {
		// uuuugh
		stack, err := activities.AwaitGetInstallStackByInstallID(ctx, sreq.ID)
		if err != nil {
			return "", errors.Wrap(err, "unable to fetch install stack")
		}
		return fmt.Sprintf("stack-%s", stack.ID), nil
	}

	if _, has := w.subwfSandbox.GetHandlers()[sreq.Type]; has {
		sandbox, err := activities.AwaitGetInstallSandboxByInstallID(ctx, sreq.ID)
		if err != nil {
			return "", errors.Wrap(err, "unable to fetch install sandbox")
		}
		return fmt.Sprintf("sandbox-%s", sandbox.ID), nil
	}

	if _, has := w.subwfActions.GetHandlers()[sreq.Type]; has {
		if sreq.InstallActionWorkflowTrigger.InstallActionWorkflowID == "" {
			panic("missing action workflow run ID")
		}
		return fmt.Sprintf("action-%s", sreq.InstallActionWorkflowTrigger.InstallActionWorkflowID), nil
	}

	if _, has := w.subwfComponents.GetHandlers()[sreq.Type]; has {
		id := sreq.ExecuteDeployComponentSubSignal.ComponentID
		if id == "" {
			id = sreq.ExecuteTeardownComponentSubSignal.ComponentID
		}
		if id == "" {
			panic("missing component ID")
		}
		comp, err := activities.AwaitGetInstallComponent(ctx, activities.GetInstallComponentRequest{
			InstallID:   sreq.ID,
			ComponentID: id,
		})
		if err != nil {
			return "", errors.Wrap(err, "unable to fetch install component")
		}
		return fmt.Sprintf("component-%s", comp.ID), nil
	}

	return "", nil
}
