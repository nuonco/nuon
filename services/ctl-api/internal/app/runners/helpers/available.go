package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
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
	job, err := s.getJob(ctx, runnerJobID)
	if err != nil {
		return fmt.Errorf("unable to get runner job: %w", err)
	}

	s.evClient.Send(ctx, job.RunnerID, &signals.Signal{
		Type:  signals.OperationJobQueued,
		JobID: runnerJobID,
	})

	return nil
}
