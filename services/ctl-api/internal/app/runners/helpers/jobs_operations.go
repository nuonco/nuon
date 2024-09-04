package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Helpers) CreateHealthCheckJob(ctx context.Context, runnerID string) (*app.RunnerJob, error) {
	job := &app.RunnerJob{
		RunnerID:          runnerID,
		QueueTimeout:      DefaultQueueTimeout,
		ExecutionTimeout:  DefaultExecutionTimeout,
		AvailableTimeout:  DefaultAvailableTimeout,
		MaxExecutions:     DefaultMaxExecutions,
		Status:            app.RunnerJobStatusQueued,
		StatusDescription: string(app.RunnerJobStatusQueued),
		Type:              app.RunnerJobTypeHealthCheck,
	}

	if res := h.db.WithContext(ctx).Create(&job); res.Error != nil {
		return nil, fmt.Errorf("unable to create job: %w", res.Error)
	}
	return job, nil
}
