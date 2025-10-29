package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/workflows"
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
		ExecFn: func(ctx workflow.Context, ereq eventloop.EventLoopRequest, sig *signals.Signal, step app.WorkflowStep) error {
			// 	sig.WorkflowStepID = step.ID
			// 	sig.WorkflowStepID = step.ID
			// 	sig.FlowStepName = step.Name
			// 	handlerSreq := signals.NewRequestSignal(ereq, sig)

			// 	// Predict the workflow ID of the underlying object's event loop
			// 	suffix, err := w.subloopSuffix(ctx, handlerSreq)
			// 	if err != nil {
			// 		return err
			// 	}

			// 	if suffix != "" {
			// 		id := fmt.Sprintf("%s-%s", sreq.ID, suffix)
			// 		if err := w.evClient.SendAndWait(ctx, id, &handlerSreq); err != nil {
			// 			return err
			// 		}
			// 	} else {
			// 		// no suffix means a workflow on this loop, so we must invoke the handler directly
			// 		handlers := w.getHandlers()
			// 		handler, ok := handlers[sig.Type]
			// 		if !ok {
			// 			return errors.New(fmt.Sprintf("no handler found for signal %s", sig.Type))
			// 		}
			// 		if err := handler(ctx, handlerSreq); err != nil {
			// 			return err
			// 		}
			// 	}

			return nil
		},
	}

	// return ufm.Handle(ctx, sreq.InstallWorkflowID)

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
		workflows.WorkflowTypeAppBranchUpdate: workflows.AppBranchUpdate,
	}
}
