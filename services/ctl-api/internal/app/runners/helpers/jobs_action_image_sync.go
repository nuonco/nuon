package helpers

import (
	"context"
	"fmt"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

// CreateActionImageSyncJob creates an oci-sync job that mirrors an
// image-backed action's app-authored image into the install registry before
// the action runs. It is owned by the action run so it is cancelled/cleaned up
// alongside it.
func (h *Helpers) CreateActionImageSyncJob(ctx context.Context,
	runnerID string,
	runID string,
	logStreamID string,
	metadata map[string]string,
) (*app.RunnerJob, error) {
	job := &app.RunnerJob{
		RunnerID:          runnerID,
		QueueTimeout:      DefaultQueueTimeout,
		ExecutionTimeout:  h.getDefaultExecutionTimeout(app.RunnerJobTypeOCISync),
		AvailableTimeout:  DefaultAvailableTimeout,
		MaxExecutions:     DefaultMaxExecutions,
		Status:            app.RunnerJobStatusQueued,
		StatusDescription: string(app.RunnerJobStatusQueued),
		Type:              app.RunnerJobTypeOCISync,
		Group:             app.RunnerJobGroupSync,
		Operation:         app.RunnerJobOperationTypeExec,
		OwnerType:         "install_action_workflow_runs",
		OwnerID:           runID,
		LogStreamID:       pkggenerics.ToPtr(logStreamID),
		Metadata:          generics.ToHstore(metadata),
	}

	if res := h.db.WithContext(ctx).Create(&job); res.Error != nil {
		return nil, fmt.Errorf("unable to create action image sync job: %w", res.Error)
	}

	return job, nil
}
