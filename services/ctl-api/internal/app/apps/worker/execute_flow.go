package worker

import (
	"encoding/json"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	queuesignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

// @temporal-gen-v2 workflow
// @execution-timeout 720h
func (w *Workflows) ExecuteFlow(ctx workflow.Context, sreq signals.RequestSignal) error {
	fc := &flow.WorkflowConductor[*signals.Signal]{
		Cfg:          w.cfg,
		V:            w.v,
		MW:           w.mw,
		Generators:   w.getWorkflowStepGenerators(ctx),
		ExecFnLegacy: w.getExecuteFlowExecFn(sreq),
	}

	err := fc.Handle(ctx, sreq.EventLoopRequest, sreq.FlowID, sreq.StartFromStepIdx)
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
		app.WorkflowTypeAppBranchesRun: workflows.AppBranchRun,
	}
}

func (w *Workflows) getExecuteFlowExecFn(sreq signals.RequestSignal) func(workflow.Context, eventloop.EventLoopRequest, *signals.Signal, app.WorkflowStep) error {
	return func(ctx workflow.Context, ereq eventloop.EventLoopRequest, sig *signals.Signal, step app.WorkflowStep) error {
		logger := workflow.GetLogger(ctx)

		// Populate signal fields from step context
		sig.FlowID = sreq.FlowID

		// Unmarshal the signal JSON from the workflow step to get the actual signal to execute
		// The step contains the serialized signal that needs to be executed
		var queueSig queuesignal.Signal
		if err := json.Unmarshal(step.Signal.SignalJSON, &queueSig); err != nil {
			return errors.Wrapf(err, "unable to unmarshal signal JSON for step %s", step.Name)
		}

		// Look up the signal constructor in the catalog to create a properly typed signal
		sigConstructor, ok := catalog.SignalCatalog[queuesignal.SignalType(step.Signal.Type)]
		if !ok {
			return errors.Errorf("signal type %s not found in catalog", step.Signal.Type)
		}

		// Create a new signal instance and unmarshal the JSON into it
		typedSignal := sigConstructor()
		if err := json.Unmarshal(step.Signal.SignalJSON, typedSignal); err != nil {
			return errors.Wrapf(err, "unable to unmarshal typed signal for step %s", step.Name)
		}

		logger.Info("enqueuing signal to queue",
			"step_name", step.Name,
			"signal_type", step.Signal.Type,
			"owner_id", ereq.ID,
			"owner_type", "app_branches",
		)

		// Enqueue the signal to the queue owned by the app branch
		// This routes the signal through the queue system for execution
		enqueueResp, err := activities.AwaitEnqueueSignalToOwner(ctx, &activities.EnqueueSignalToOwnerRequest{
			OwnerID:   ereq.ID, // App branch ID
			OwnerType: "app_branches",
			Signal:    typedSignal,
		})
		if err != nil {
			return errors.Wrapf(err, "unable to enqueue signal for step %s", step.Name)
		}

		logger.Info("waiting for queue signal to complete",
			"step_name", step.Name,
			"queue_signal_id", enqueueResp.QueueSignalID,
			"workflow_id", enqueueResp.WorkflowID,
		)

		// Wait for the queue signal to complete execution
		_, err = client.AwaitAwaitSignal(ctx, enqueueResp.QueueSignalID)
		if err != nil {
			return errors.Wrapf(err, "queue signal execution failed for step %s", step.Name)
		}

		logger.Info("queue signal completed successfully", "step_name", step.Name)

		return nil
	}
}
