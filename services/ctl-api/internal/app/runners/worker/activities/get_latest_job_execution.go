package activities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetLatestJobExecutionRequest struct {
	JobID       string    `validate:"required"`
	AvailableAt time.Time `validate:"required"`
}

type GetLatestJobExecutionResponse struct {
	Found        bool
	JobExecution *app.RunnerJobExecution
}

// @temporal-gen activity
func (a *Activities) GetLatestJobExecution(ctx context.Context, req GetLatestJobExecutionRequest) (*GetLatestJobExecutionResponse, error) {
	jobExecution, err := a.getLatestJobExecution(ctx, req.JobID, req.AvailableAt)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &GetLatestJobExecutionResponse{
				Found: false,
			}, nil
		}

		return nil, fmt.Errorf("unable to get runner job execution: %w", err)
	}

	return &GetLatestJobExecutionResponse{
		Found:        true,
		JobExecution: jobExecution,
	}, nil
}

func (a *Activities) getLatestJobExecution(ctx context.Context, jobID string, availableAt time.Time) (*app.RunnerJobExecution, error) {
	jobExecution := app.RunnerJobExecution{}
	res := a.db.WithContext(ctx).
		Where(app.RunnerJobExecution{
			RunnerJobID: jobID,
		}).
		Order("created_at desc").
		Limit(1).
		First(&jobExecution)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get latest job execution: %w", res.Error)
	}

	if jobExecution.CreatedAt.Before(availableAt) {
		return nil, gorm.ErrRecordNotFound
	}

	return &jobExecution, nil
}
