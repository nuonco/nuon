package rerunflow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
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

// rerunFlow runs the workflow conductor's Rerun with v2 generators and queue-based execution.
// Handles ContinueAsNewErr by looping internally (same rationale as executeFlow).
func (s *Signal) rerunFlow(ctx workflow.Context) error {
	eventLoopReq := eventloop.EventLoopRequest{
		ID: s.InstallID,
	}

	// Check if steps-workflows feature is enabled
	stepsWorkflows, _ := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureStepsWorkflows))

	fc := &flow.WorkflowConductor[*signals.Signal]{
		Generators:          getWorkflowStepGenerators(),
		StepChildWorkflow:   stepsWorkflows,
		StepQueueName:       "install-workflow-steps",
		StepTargetQueueName: "install-signals",
		StepOwnerID:         s.InstallID,
		StepOwnerType:       "installs",
	}

	continueFromIdx := 0
	for {
		err := fc.Rerun(ctx, eventLoopReq, flow.RerunInput{
			ContinueFromIdx: continueFromIdx,
			FlowID:          s.InstallWorkflowID,
			StepID:          s.RerunConfiguration.StepID,
			StalePlan:       s.RerunConfiguration.StalePlan,
			RePlanStepID:    s.RerunConfiguration.RePlanStepID,
			Operation:       flow.RerunOperation(s.RerunConfiguration.StepOperation),
		})
		if err == nil {
			return nil
		}

		cerr, ok := err.(*flow.ContinueAsNewErr)
		if ok && cerr != nil {
			continueFromIdx = cerr.StartFromStepIdx
			continue
		}

		return err
	}
}
