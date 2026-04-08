package executeflow

import (
	pkgerrors "github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	v2workflows "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/workflows/v2"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

func getWorkflowStepGenerators() map[app.WorkflowType]flow.WorkflowStepGenerator {
	return map[app.WorkflowType]flow.WorkflowStepGenerator{
		app.WorkflowTypeManualDeploy:               v2workflows.ManualDeploySteps,
		app.WorkflowTypeDriftRun:                   v2workflows.ManualDeploySteps,
		app.WorkflowTypeDeployComponents:           v2workflows.DeployAllComponents,
		app.WorkflowTypeTeardownComponent:          v2workflows.TeardownComponent,
		app.WorkflowTypeTeardownComponents:         v2workflows.TeardownComponents,
		app.WorkflowTypeInputUpdate:                v2workflows.InputUpdate,
		app.WorkflowTypeActionWorkflowRun:          v2workflows.RunActionWorkflow,
		app.WorkflowTypeProvision:                  v2workflows.Provision,
		app.WorkflowTypeReprovision:                v2workflows.Reprovision,
		app.WorkflowTypeReprovisionSandbox:         v2workflows.ReprovisionSandbox,
		app.WorkflowTypeDriftRunReprovisionSandbox: v2workflows.ReprovisionSandbox,
		app.WorkflowTypeDeprovision:                v2workflows.Deprovision,
		app.WorkflowTypeDeprovisionSandbox:         v2workflows.DeprovisionSandbox,
		app.WorkflowTypeSyncSecrets:                v2workflows.SyncSecrets,
	}
}

func getExecuteFlowExecFn(installID string) func(workflow.Context, *signaldb.SignalData, app.WorkflowStep) error {
	return func(ctx workflow.Context, queueSignal *signaldb.SignalData, step app.WorkflowStep) error {
		logger := workflow.GetLogger(ctx)

		sig := queueSignal.Signal
		if sig == nil {
			logger.Info("enqueing nil signals to install queue",
				"step_name", step.Name,
				"owner_id", installID,
				"owner_type", "installs",
			)
			return nil
		}

		logger.Info("enqueuing signal to install queue",
			"step_name", step.Name,
			"step_id", step.ID,
			"owner_id", installID,
			"owner_type", "installs",
		)

		enqueueResp, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:         installID,
			OwnerType:       "installs",
			QueueName:       "install-signals",
			Signal:          sig,
			SignalOwnerID:   step.ID,
			SignalOwnerType: "install_workflow_steps",
		})
		if err != nil {
			return pkgerrors.Wrapf(err, "unable to enqueue signal for step %s", step.Name)
		}

		logger.Info("waiting for queue signal to complete",
			"step_name", step.Name,
			"queue_signal_id", enqueueResp.QueueSignalID,
		)

		_, err = client.AwaitAwaitSignal(ctx, enqueueResp.QueueSignalID)
		if err != nil {
			return pkgerrors.Wrapf(err, "queue signal execution failed for step %s", step.Name)
		}

		logger.Info("queue signal completed successfully", "step_name", step.Name)

		return nil
	}
}

// executeFlow runs the workflow conductor with v2 generators and queue-based execution.
// It handles ContinueAsNewErr by looping internally since the signal Execute() runs
// inside a queue handler workflow and cannot use Temporal's ContinueAsNew directly.
// This is safe because each step is just enqueue+await (2 lightweight activity calls).
func (s *Signal) executeFlow(ctx workflow.Context) error {
	eventLoopReq := eventloop.EventLoopRequest{
		ID: s.installID,
	}

	fc := &flow.WorkflowConductor[*signals.Signal]{
		Generators:        getWorkflowStepGenerators(),
		ExecFn:            getExecuteFlowExecFn(s.installID),
		StepChildWorkflow: false,
	}

	startFromStepIdx := 0
	for {
		err := fc.Handle(ctx, eventLoopReq, s.InstallWorkflowID, startFromStepIdx)
		if err == nil {
			return nil
		}

		cerr, ok := err.(*flow.ContinueAsNewErr)
		if ok && cerr != nil {
			startFromStepIdx = cerr.StartFromStepIdx
			continue
		}

		return err
	}
}
