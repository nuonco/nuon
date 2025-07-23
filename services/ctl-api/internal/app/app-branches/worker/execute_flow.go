package worker

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/app-branches/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/app-branches/workflows"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/flow"
)

// @temporal-gen workflow
// @execution-timeout 720h
func (w *Workflows) ExecuteFlow(ctx workflow.Context, sreq signals.RequestSignal) error {
	fc := &flow.WorkflowConductor[*signals.Signal]{
		Cfg:        w.cfg,
		V:          w.v,
		MW:         w.mw,
		Generators: w.getWorkflowStepGenerators(ctx),
		ExecFn:     w.getExecuteFlowExecFn(sreq),
	}

	return fc.Handle(ctx, sreq.EventLoopRequest, sreq.FlowID)
}

func (w *Workflows) getWorkflowStepGenerators(ctx workflow.Context) map[app.WorkflowType]flow.WorkflowStepGenerator {
	return map[app.WorkflowType]flow.WorkflowStepGenerator{
		app.WorkflowTypeAppBranchesManualUpdate:        workflows.ManualUpdateSteps,
		app.WorkflowTypeAppBranchesComponentRepoUpdate: workflows.ComponentRepoUpdateSteps,
		app.WorkflowTypeAppBranchesConfigRepoUpdate:    workflows.ConfigRepoUpdateSteps,
	}
}

func (w *Workflows) getExecuteFlowExecFn(sreq signals.RequestSignal) func(workflow.Context, eventloop.EventLoopRequest, *signals.Signal, app.WorkflowStep) error {
	return func(ctx workflow.Context, ereq eventloop.EventLoopRequest, sig *signals.Signal, step app.WorkflowStep) error {
		sig.FlowID = sreq.FlowID
		sig.WorkflowStepID = step.ID
		sig.WorkflowStepName = step.Name

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

	return "", nil
}
