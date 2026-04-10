package executeflow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	v2workflows "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/workflows/v2"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow"
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

// executeFlow runs the workflow conductor with v2 generators and queue-based execution.
// It handles ContinueAsNewErr by looping internally since the signal Execute() runs
// inside a queue handler workflow and cannot use Temporal's ContinueAsNew directly.
// This is safe because each step is just enqueue+await (2 lightweight activity calls).
func (s *Signal) executeFlow(ctx workflow.Context) error {
	eventLoopReq := eventloop.EventLoopRequest{
		ID: s.installID,
	}

	fc := &flow.WorkflowConductor[*signals.Signal]{
		Generators:          getWorkflowStepGenerators(),
		StepChildWorkflow:   true,
		StepQueueName:       "install-workflow-steps",
		StepTargetQueueName: "install-signals",
		StepOwnerID:         s.installID,
		StepOwnerType:       "installs",
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
