package helpers

import (
	"context"
	"fmt"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

func (h *Helpers) CreateRunnerJob(ctx context.Context,
	runnerID string,
	ownerType string,
	ownerID string,
	typ app.RunnerJobType,
	op app.RunnerJobOperationType,
	logStreamID string,
	metadata map[string]string,
) (*app.RunnerJob, error) {
	job := &app.RunnerJob{
		RunnerID:          runnerID,
		OwnerType:         ownerType,
		OwnerID:           ownerID,
		QueueTimeout:      DefaultQueueTimeout,
		ExecutionTimeout:  h.getDefaultExecutionTimeout(typ),
		AvailableTimeout:  DefaultAvailableTimeout,
		MaxExecutions:     DefaultMaxExecutions,
		Status:            app.RunnerJobStatusQueued,
		StatusDescription: string(app.RunnerJobStatusQueued),
		Type:              typ,
		Operation:         op,
		LogStreamID:       pkggenerics.ToPtr(logStreamID),
		Metadata:          generics.ToHstore(metadata),
	}

	if err := h.createJobWithCredentials(ctx, job); err != nil {
		return nil, err
	}

	return job, nil
}

func (h *Helpers) GetRunnerJob(ctx context.Context, jobID string) (*app.RunnerJob, error) {
	job := &app.RunnerJob{}
	if res := h.db.WithContext(ctx).Where("id = ?", jobID).First(&job); res.Error != nil {
		return nil, fmt.Errorf("unable to get runner job: %w", res.Error)
	}

	return job, nil
}

func (h *Helpers) CreateMngRunnerJob(
	ctx context.Context,
	runnerID string,
	typ app.RunnerJobType,
	op app.RunnerJobOperationType,
	logStreamID string,
	metadata map[string]string,
) (*app.RunnerJob, error) {
	job := &app.RunnerJob{
		RunnerID:          runnerID,
		OwnerType:         "runners",
		OwnerID:           runnerID,
		QueueTimeout:      DefaultQueueTimeout,
		ExecutionTimeout:  h.getDefaultExecutionTimeout(typ),
		AvailableTimeout:  DefaultAvailableTimeout,
		MaxExecutions:     DefaultMaxExecutions,
		Status:            app.RunnerJobStatusQueued,
		StatusDescription: string(app.RunnerJobStatusQueued),
		Type:              typ,
		Operation:         op,
		Group:             app.RunnerJobGroupManagement,
		LogStreamID:       pkggenerics.ToPtr(logStreamID),
		Metadata:          generics.ToHstore(metadata),
	}

	if err := h.createJobWithCredentials(ctx, job); err != nil {
		return nil, err
	}

	return job, nil
}

func (h *Helpers) createJobWithCredentials(ctx context.Context, job *app.RunnerJob) error {
	return account.New(account.Params{DB: h.db}).CreateRunnerJob(ctx, job)
}
