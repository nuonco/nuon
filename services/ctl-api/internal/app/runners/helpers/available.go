package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *Helpers) MarkJobAvailable(ctx context.Context, runnerJobID string) error {
	runnerJob := app.RunnerJob{
		ID: runnerJobID,
	}

	res := s.db.WithContext(ctx).Model(&runnerJob).Updates(app.RunnerJob{
		Status:            app.RunnerJobStatusAvailable,
		StatusDescription: string(app.RunnerJobStatusAvailable),
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update job status: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no job found: %s %w", runnerJobID, gorm.ErrRecordNotFound)
	}

	return nil
}

func (s *Helpers) QueueJob(ctx context.Context, runnerJobID string) error {
	_, err := s.getJob(ctx, runnerJobID)
	if err != nil {
		return fmt.Errorf("unable to get runner job: %w", err)
	}

	return nil
}
