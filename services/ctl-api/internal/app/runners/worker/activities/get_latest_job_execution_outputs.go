package activities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetLatestJobExecutionOutputsRequest struct {
	JobID       string    `validate:"required"`
	AvailableAt time.Time `validate:"required"`
}

type GetLatestJobExecutionOutputsResponse struct {
	Found        bool                    `validate:"required"`
	JobExecution *app.RunnerJobExecution `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) GetLatestJobExecutionOutputs(ctx context.Context, req GetLatestJobExecutionOutputsRequest) (*GetLatestJobExecutionOutputsResponse, error) {
	jobExecution, err := a.getLatestJobExecutionWithOutputs(ctx, req.JobID, req.AvailableAt)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &GetLatestJobExecutionOutputsResponse{
				Found: false,
			}, nil
		}

		return nil, fmt.Errorf("unable to get runner job execution: %w", err)
	}

	return &GetLatestJobExecutionOutputsResponse{
		Found:        true,
		JobExecution: jobExecution,
	}, nil
}

func (a *Activities) getLatestJobExecutionWithOutputs(ctx context.Context, jobID string, availableAt time.Time) (*app.RunnerJobExecution, error) {
	jobExecution := app.RunnerJobExecution{}
	res := a.db.WithContext(ctx).
		Where(app.RunnerJobExecution{
			RunnerJobID: jobID,
		}).
		Order("created_at desc").
		Preload("Outputs").
		Limit(1).
		First(&jobExecution)
	if res.Error != nil {
	}

	if jobExecution.CreatedAt.Before(availableAt) {
		return nil, gorm.ErrRecordNotFound
	}

	return &jobExecution, nil
}
