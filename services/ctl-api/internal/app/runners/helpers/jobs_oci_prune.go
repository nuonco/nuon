package helpers

import (
	"context"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

func (h *Helpers) CreateOCIPruneJob(ctx context.Context,
	runnerID string,
	installID string,
	logStreamID string,
	metadata map[string]string,
) (*app.RunnerJob, error) {
	job := &app.RunnerJob{
		RunnerID:          runnerID,
		QueueTimeout:      DefaultQueueTimeout,
		ExecutionTimeout:  h.getDefaultExecutionTimeout(app.RunnerJobTypeOCIPrune),
		AvailableTimeout:  DefaultAvailableTimeout,
		MaxExecutions:     1,
		Status:            app.RunnerJobStatusQueued,
		StatusDescription: string(app.RunnerJobStatusQueued),
		Type:              app.RunnerJobTypeOCIPrune,
		Group:             app.RunnerJobGroupSync,
		Operation:         app.RunnerJobOperationTypeExec,
		OwnerType:         "installs",
		OwnerID:           installID,
		LogStreamID:       pkggenerics.ToPtr(logStreamID),
		Metadata:          generics.ToHstore(metadata),
	}

	if err := h.createJobWithCredentials(ctx, job); err != nil {
		return nil, err
	}

	return job, nil
}
