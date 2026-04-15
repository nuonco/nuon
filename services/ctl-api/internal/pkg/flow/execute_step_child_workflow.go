package flow

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// dispatchFlowStepSignal delegates to the package-level DispatchStepSignal.
func (c *WorkflowConductor[DomainSignal]) dispatchFlowStepSignal(ctx workflow.Context, step *app.WorkflowStep, flw *app.Workflow) error {
	return DispatchStepSignal(ctx, c.stepConfig(), step, flw)
}

// stepConfig returns a StepConfig from the conductor's fields.
func (c *WorkflowConductor[DomainSignal]) stepConfig() StepConfig {
	return StepConfig{
		QueueName:       c.StepQueueName,
		TargetQueueName: c.StepTargetQueueName,
		OwnerID:         c.StepOwnerID,
		OwnerType:       c.StepOwnerType,
	}
}
