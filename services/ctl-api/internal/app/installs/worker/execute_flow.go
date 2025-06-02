package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/flows"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/flow"
)

// @temporal-gen workflow
// @execution-timeout 720h
func (w *Workflows) ExecuteFlow(ctx workflow.Context, sreq signals.RequestSignal) error {
	if sreq.FlowID == "" {
		sreq.FlowID = sreq.InstallWorkflowID
	}
	ufm := &flow.FlowConductor[*signals.Signal]{
		Cfg:        w.cfg,
		V:          w.v,
		MW:         w.mw,
		Generators: w.getFlowStepGenerators(ctx),
		ExecFn: func(ctx workflow.Context, ereq eventloop.EventLoopRequest, sig *signals.Signal, step app.FlowStep) error {
			sig.WorkflowStepID = step.ID
			sig.WorkflowStepID = step.ID
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
		},
	}

	// return ufm.Handle(ctx, sreq.InstallWorkflowID)
	return ufm.Handle(ctx, sreq.EventLoopRequest, sreq.FlowID)
}

func (w *Workflows) getFlowStepGenerators(ctx workflow.Context) map[app.FlowType]flow.FlowStepGenerator {
	ufg := flows.NewFlows(flows.Params{
		Cfg: w.cfg,
		MW:  w.mw,
		V:   w.v,
	})

	return map[app.FlowType]flow.FlowStepGenerator{
		flows.FlowTypeManualDeploy:       ufg.ManualDeploySteps,
		flows.FlowTypeDeployComponents:   ufg.DeployAllComponents,
		flows.FlowTypeTeardownComponent:  ufg.TeardownComponent,
		flows.FlowTypeTeardownComponents: ufg.TeardownComponents,
		flows.FlowTypeInputUpdate:        ufg.InputUpdate,
		flows.FlowTypeActionWorkflowRun:  ufg.RunActionWorkflow,
		flows.FlowTypeProvision:          ufg.Provision,
		flows.FlowTypeReprovision:        ufg.Reprovision,
		flows.FlowTypeReprovisionSandbox: ufg.ReprovisionSandbox,
		flows.FlowTypeDeprovision:        ufg.Deprovision,
		flows.FlowTypeDeprovisionSandbox: ufg.DeprovisionSandbox,
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
