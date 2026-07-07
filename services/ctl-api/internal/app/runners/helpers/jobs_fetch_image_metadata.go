package helpers

import (
	"context"
	"fmt"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

// CreateFetchImageMetadataJob creates a job for fetching image metadata from an OCI registry.
// This job is used during external image builds to fetch metadata for policy evaluation.
func (h *Helpers) CreateFetchImageMetadataJob(ctx context.Context,
	runnerID string,
	executor app.RunnerJobExecutor,
	ownerType string,
	ownerID string,
	logStreamID string,
	metadata map[string]string,
) (*app.RunnerJob, error) {
	job := &app.RunnerJob{
		RunnerID:          runnerID,
		Executor:          executor,
		OwnerType:         ownerType,
		OwnerID:           ownerID,
		QueueTimeout:      DefaultQueueTimeout,
		ExecutionTimeout:  h.getDefaultExecutionTimeout(app.RunnerJobTypeFetchImageMetadata),
		AvailableTimeout:  DefaultAvailableTimeout,
		MaxExecutions:     DefaultMaxExecutions,
		Status:            app.RunnerJobStatusQueued,
		StatusDescription: string(app.RunnerJobStatusQueued),
		Type:              app.RunnerJobTypeFetchImageMetadata,
		Group:             app.RunnerJobGroupSync,
		Operation:         app.RunnerJobOperationTypeExec,
		LogStreamID:       pkggenerics.ToPtr(logStreamID),
		Metadata:          generics.ToHstore(metadata),
	}

	db := h.db.WithContext(ctx)
	if executor == app.RunnerJobExecutorControlPlane {
		db = db.Omit("RunnerID")
	}
	if res := db.Create(&job); res.Error != nil {
		return nil, fmt.Errorf("unable to create job: %w", res.Error)
	}

	return job, nil
}
