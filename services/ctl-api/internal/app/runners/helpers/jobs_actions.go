package helpers

import (
	"context"
	"fmt"
	"time"

	pkggenerics "github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

func (h *Helpers) CreateActionsWorkflowRunJob(ctx context.Context,
	runnerID string,
	runID string, logStreamID string,
	cfg *app.ActionWorkflowConfig,
	metadata map[string]string,
) (*app.RunnerJob, error) {
	job := &app.RunnerJob{
		RunnerID:          runnerID,
		QueueTimeout:      DefaultQueueTimeout,
                // NOTE(jm): we create a buffer to allow the runner to finish cleaning up, when a job times out. If this 
                // is set to the exact timeout, the workflow can not actually cleanup. Thus, this is an arbitrary buffer 
                // that should never be hit because the runner job would timeout mid-job and mark itself as timed out 
                // first.
		ExecutionTimeout:  cfg.Timeout + time.Minute,
		AvailableTimeout:  DefaultAvailableTimeout,
		MaxExecutions:     DefaultMaxExecutions,
		Status:            app.RunnerJobStatusQueued,
		StatusDescription: string(app.RunnerJobStatusQueued),
		Type:              app.RunnerJobTypeActionsWorkflowRun,
		Group:             app.RunnerJobGroupActions,
		Operation:         app.RunnerJobOperationTypeExec,
		OwnerType:         "install_action_workflow_runs",
		OwnerID:           runID,
		LogStreamID:       pkggenerics.ToPtr(logStreamID),
		Metadata:          generics.ToHstore(metadata),
	}

	if res := h.db.WithContext(ctx).Create(&job); res.Error != nil {
		return nil, fmt.Errorf("unable to create job: %w", res.Error)
	}

	return job, nil
}
