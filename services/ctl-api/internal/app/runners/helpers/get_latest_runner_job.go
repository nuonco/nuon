package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type GetLatestJobRequest struct {
	OwnerID string
}

func (a *Helpers) GetLatestJob(ctx context.Context, req *GetLatestJobRequest) (*app.RunnerJob, error) {
	var job app.RunnerJob
	res := a.db.WithContext(ctx).
		Where(app.RunnerJob{
			OwnerID: req.OwnerID,
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
