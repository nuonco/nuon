package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateSandboxJobExecutionRequest struct {
	JobID string `validate:"required" temporaljson:"job_id,omitzero,omitempty"`
}

// @temporal-gen-v2 activity
// @max-retries 3
func (a *Activities) CreateSandboxJobExecution(ctx context.Context, req CreateSandboxJobExecutionRequest) error {
	var job app.RunnerJob
	if res := a.db.WithContext(ctx).Where(app.RunnerJob{ID: req.JobID}).First(&job); res.Error != nil {
		return fmt.Errorf("unable to get runner job %s: %w", req.JobID, res.Error)
	}

	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		execution := app.RunnerJobExecution{
			RunnerJobID: job.ID,
			Status:      app.RunnerJobExecutionStatusFinished,
			OrgID:       job.OrgID,
			CreatedByID: job.CreatedByID,
		}
		if err := tx.Create(&execution).Error; err != nil {
			return fmt.Errorf("unable to create sandbox job execution: %w", err)
		}

		result := app.RunnerJobExecutionResult{
			RunnerJobExecutionID: execution.ID,
			Success:              true,
			OrgID:                job.OrgID,
			CreatedByID:          job.CreatedByID,
		}
		if err := tx.Create(&result).Error; err != nil {
			return fmt.Errorf("unable to create sandbox job execution result: %w", err)
		}

		return nil
	})
}
