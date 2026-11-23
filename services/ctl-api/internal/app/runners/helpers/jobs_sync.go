package helpers

import (
	"context"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

func (h *Helpers) CreateSyncJob(ctx context.Context,
	runnerID string,
	typ app.RunnerJobType,
	op app.RunnerJobOperationType,
	deployID string,
	logStreamID string,
	metadata map[string]string,
) (*app.RunnerJob, error) {
	job := &app.RunnerJob{
		RunnerID:          runnerID,
		QueueTimeout:      DefaultQueueTimeout,
		ExecutionTimeout:  h.getDefaultExecutionTimeout(typ),
		AvailableTimeout:  DefaultAvailableTimeout,
		MaxExecutions:     DefaultMaxExecutions,
		Status:            app.RunnerJobStatusQueued,
		StatusDescription: string(app.RunnerJobStatusQueued),
		Type:              typ,
		Group:             app.RunnerJobGroupSync,
		Operation:         op,
		OwnerType:         "install_deploys",
		OwnerID:           deployID,
		LogStreamID:       pkggenerics.ToPtr(logStreamID),
		Metadata:          generics.ToHstore(metadata),
	}

	if err := h.createJobWithCredentials(ctx, job); err != nil {
		return nil, err
	}

	return job, nil
}
