package activities

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type WriteSandboxJobOutputsRequest struct {
	JobID string `validate:"required"`
}

// WriteSandboxJobOutputs materialises the fake outputs embedded in a sandbox
// job's plan into RunnerJobExecution + RunnerJobExecutionOutputs rows so that
// downstream code (e.g. shared_execute_sync) sees them via the
// runner_jobs_view_v2 → RunnerJob.ParsedOutputs path.
//
// This is the ctl-api-side equivalent of the runner's sandboxOutputs() path
// (job_step_outputs.go) — same input (plan JSON), same output (execution
// outputs row), just written directly when the runner is not running.
//
// Follows the SyncNoopDeployOutputs pattern
// (pkg/workflows/workflow/activities/sync_noop_deploy_outputs.go).
//
// @temporal-gen-v2 activity
// @by-field JobID
func (a *Activities) WriteSandboxJobOutputs(ctx context.Context, req *WriteSandboxJobOutputsRequest) error {
	var plan app.RunnerJobPlan
	if err := a.db.WithContext(ctx).
		Where("runner_job_id = ?", req.JobID).
		First(&plan).Error; err != nil {
		// No plan yet — nothing to write. This is not an error; some jobs
		// (e.g. healthchecks) don't have plans.
		return nil
	}

	var min plantypes.MinSandboxMode
	if err := json.Unmarshal([]byte(plan.PlanJSON), &min); err != nil {
		return nil
	}

	if min.SandboxMode == nil || !min.SandboxMode.Enabled || len(min.SandboxMode.Outputs) == 0 {
		return nil
	}

	outputsJSON, err := json.Marshal(min.SandboxMode.Outputs)
	if err != nil {
		return fmt.Errorf("unable to marshal sandbox outputs: %w", err)
	}

	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		execution := &app.RunnerJobExecution{
			RunnerJobID: req.JobID,
			OrgID:       plan.OrgID,
			CreatedByID: plan.CreatedByID,
			Status:      app.RunnerJobExecutionStatusFinished,
		}
		if err := tx.Create(execution).Error; err != nil {
			return fmt.Errorf("unable to create sandbox execution: %w", err)
		}

		executionOutputs := &app.RunnerJobExecutionOutputs{
			RunnerJobExecutionID: execution.ID,
			OrgID:                plan.OrgID,
			CreatedByID:          plan.CreatedByID,
			Outputs:              outputsJSON,
		}
		if err := tx.Create(executionOutputs).Error; err != nil {
			return fmt.Errorf("unable to create sandbox execution outputs: %w", err)
		}

		return nil
	})
}
