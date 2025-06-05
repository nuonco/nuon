package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetJobRequest struct {
	ID string
}

// @temporal-gen activity
// @by-id ID
func (a *Activities) GetJob(ctx context.Context, req *GetJobRequest) (*app.RunnerJob, error) {
	job := app.RunnerJob{}
	res := a.db.WithContext(ctx).
		Preload("Executions").
		Preload("Executions.Result", func(db *gorm.DB) *gorm.DB {
			return db.Order("runner_job_execution_results.created_at DESC")
		}).
		First(&job, "id = ?", req.ID)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get job: %w", res.Error)
	}

	return &job, nil
}
