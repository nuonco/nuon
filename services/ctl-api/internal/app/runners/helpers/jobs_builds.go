package helpers

import (
	"context"
	"fmt"
	"time"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

func (h *Helpers) CreateBuildJob(ctx context.Context,
	runnerID string,
	ownerType string,
	ownerID string,
	typ app.RunnerJobType,
	op app.RunnerJobOperationType,
	logStreamID string,
	metadata map[string]string,
	customTimeout *time.Duration,
) (*app.RunnerJob, error) {
	timeout := h.getDefaultExecutionTimeout(typ)
	if customTimeout != nil && *customTimeout > 0 {
		timeout = *customTimeout
		if timeout < app.MinBuildTimeout {
			timeout = app.MinBuildTimeout
		}
		if timeout > app.MaxBuildTimeout {
			timeout = app.MaxBuildTimeout
		}
	}

	job := &app.RunnerJob{
		RunnerID:          runnerID,
		OwnerType:         ownerType,
		OwnerID:           ownerID,
		QueueTimeout:      DefaultQueueTimeout,
		ExecutionTimeout:  timeout,
		AvailableTimeout:  DefaultAvailableTimeout,
		MaxExecutions:     DefaultMaxExecutions,
		Status:            app.RunnerJobStatusQueued,
		StatusDescription: string(app.RunnerJobStatusQueued),
		Group:             app.RunnerJobGroupBuild,
		Type:              typ,
		Operation:         op,
		LogStreamID:       pkggenerics.ToPtr(logStreamID),
		Metadata:          generics.ToHstore(metadata),
	}

	if res := h.db.WithContext(ctx).Create(&job); res.Error != nil {
		return nil, fmt.Errorf("unable to create job: %w", res.Error)
	}

	return job, nil
}
