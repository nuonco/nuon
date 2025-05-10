package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

func (h *Helpers) CreateSyncSecretsJob(ctx context.Context,
	runnerID string,
	ownerType string,
	ownerID string,
	typ app.RunnerJobType,
	op app.RunnerJobOperationType,
	metadata map[string]string,
) (*app.RunnerJob, error) {
	job := &app.RunnerJob{
		RunnerID:          runnerID,
		OwnerType:         ownerType,
		OwnerID:           ownerID,
		QueueTimeout:      DefaultQueueTimeout,
		ExecutionTimeout:  h.getExecutionTimeout(typ),
		AvailableTimeout:  DefaultAvailableTimeout,
		MaxExecutions:     DefaultMaxExecutions,
		Status:            app.RunnerJobStatusQueued,
		StatusDescription: string(app.RunnerJobStatusQueued),
		Type:              typ,
		Operation:         op,
		Metadata:          generics.ToHstore(metadata),
	}

	if res := h.db.WithContext(ctx).Create(&job); res.Error != nil {
		return nil, fmt.Errorf("unable to create job: %w", res.Error)
	}

	return job, nil
}
