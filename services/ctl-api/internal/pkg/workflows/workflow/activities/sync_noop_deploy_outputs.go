package activities

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type SyncNoopDeployOutputsRequest struct {
	StepID string `validate:"required"`
}

// SyncNoopDeployOutputs copies outputs from the latest plan job to a new
// finished apply job on the deploy. This is called when a deploy plan is noop
// (infrastructure matches desired state) so that the deploy has outputs just
// like a normal apply would produce. The plan job already collected outputs
// via terraform output / helm get values.
//
// @temporal-gen-v2 activity
// @max-retries 1
func (a *Activities) SyncNoopDeployOutputs(ctx context.Context, req *SyncNoopDeployOutputsRequest) error {
	step, err := a.PkgWorkflowsFlowGetFlowsStep(ctx, GetFlowStepRequest{
		FlowStepID: req.StepID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to get step")
	}

	if step.StepTargetType != "install_deploys" {
		return nil
	}

	// Get the latest plan job (uses the view, so outputs are available).
	var planJob app.RunnerJob
	err = a.db.WithContext(ctx).
		Where("owner_id = ? AND operation = ?", step.StepTargetID, app.RunnerJobOperationTypeCreateApplyPlan).
		Order("created_at DESC").
		First(&planJob).Error
	if err != nil {
		return nil
	}

	if len(planJob.ParsedOutputs) == 0 {
		return nil
	}

	outputsJSON, err := json.Marshal(planJob.ParsedOutputs)
	if err != nil {
		return nil
	}

	// Create a finished apply job with the plan's outputs so the deploy
	// has outputs in the same structure the rest of the system expects.
	applyJob := &app.RunnerJob{
		RunnerID:          planJob.RunnerID,
		OrgID:             planJob.OrgID,
		OwnerID:           planJob.OwnerID,
		OwnerType:         planJob.OwnerType,
		Status:            app.RunnerJobStatusFinished,
		StatusDescription: "noop apply - outputs synced from plan",
		Type:              planJob.Type,
		Group:             planJob.Group,
		Operation:         app.RunnerJobOperationTypeApplyPlan,
		QueueTimeout:      planJob.QueueTimeout,
		AvailableTimeout:  planJob.AvailableTimeout,
		ExecutionTimeout:  planJob.ExecutionTimeout,
		MaxExecutions:     1,
	}

	if err := a.db.WithContext(ctx).Create(applyJob).Error; err != nil {
		return errors.Wrap(err, "unable to create noop apply job")
	}

	execution := &app.RunnerJobExecution{
		RunnerJobID: applyJob.ID,
		Status:      app.RunnerJobExecutionStatusFinished,
	}
	if err := a.db.WithContext(ctx).Create(execution).Error; err != nil {
		return errors.Wrap(err, "unable to create noop apply execution")
	}

	executionOutputs := &app.RunnerJobExecutionOutputs{
		RunnerJobExecutionID: execution.ID,
		OrgID:                planJob.OrgID,
		Outputs:              outputsJSON,
	}
	if err := a.db.WithContext(ctx).Create(executionOutputs).Error; err != nil {
		return errors.Wrap(err, "unable to create noop apply outputs")
	}

	return nil
}
