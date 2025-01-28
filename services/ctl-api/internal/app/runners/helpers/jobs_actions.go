package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (h *Helpers) CreateActionsWorkflowRunJob(ctx context.Context,
	runnerID string,
	runID string, logStreamID string,
	cfg *app.ActionWorkflowConfig,
) (*app.RunnerJob, error) {
	job := &app.RunnerJob{
		RunnerID:          runnerID,
		QueueTimeout:      DefaultQueueTimeout,
		ExecutionTimeout:  cfg.Timeout,
		AvailableTimeout:  DefaultAvailableTimeout,
		MaxExecutions:     DefaultMaxExecutions,
		Status:            app.RunnerJobStatusQueued,
		StatusDescription: string(app.RunnerJobStatusQueued),
		Type:              app.RunnerJobTypeActionsWorkflowRun,
		Group:             app.RunnerJobGroupActions,
		Operation:         app.RunnerJobOperationTypeExec,
		OwnerType:         "install_action_workflow_runs",
		OwnerID:           runID,
		LogStreamID:       generics.ToPtr(logStreamID),
	}

	if res := h.db.WithContext(ctx).Create(&job); res.Error != nil {
		return nil, fmt.Errorf("unable to create job: %w", res.Error)
	}

	return job, nil
}
