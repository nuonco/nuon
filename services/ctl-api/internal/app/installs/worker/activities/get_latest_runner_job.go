package activities

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetLatestJobRequest struct {
	OwnerID   string                     `validate:"required"`
	Operation app.RunnerJobOperationType `validate:"required"`
	Group     app.RunnerJobGroup         `validate:"required"`
	Type      app.RunnerJobType          `validate:"required"`
}

// @temporal-gen activity
// @by-id OwnerID
func (a *Activities) GetLatestJob(ctx context.Context, req *GetLatestJobRequest) (*app.RunnerJob, error) {
	job := app.RunnerJob{}
	res := a.db.WithContext(ctx).
		Where(app.RunnerJob{
			OwnerID:   req.OwnerID,
			Group:     req.Group,
			Type:      req.Type,
			Operation: req.Operation,
		}).
		Preload("Executions").
		Preload("Executions.Result", func(db *gorm.DB) *gorm.DB {
			return db.Order("runner_job_execution_results.created_at DESC")
		}).
		Order("created_at desc").
		Limit(1).
		First(&job)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get job: %w", res.Error)
	}

	return &job, nil
}
