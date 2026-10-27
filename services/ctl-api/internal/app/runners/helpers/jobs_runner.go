package helpers

import (
	"context"
	"fmt"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
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

	if res := h.db.WithContext(ctx).Create(&job); res.Error != nil {
		return nil, fmt.Errorf("unable to create job: %w", res.Error)
	}
	if err := h.provisionJobCredentials(ctx, job); err != nil {
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

	if res := h.db.WithContext(ctx).Create(&job); res.Error != nil {
		return nil, fmt.Errorf("unable to create job: %w", res.Error)
	}
	if err := h.provisionJobCredentials(ctx, job); err != nil {
		return nil, err
	}

	return job, nil
}

func (h *Helpers) provisionJobCredentials(ctx context.Context, job *app.RunnerJob) error {
	orgID, err := orgIDForJob(ctx, h, job)
	if err != nil {
		return err
	}
	if err := h.ProvisionJobServiceAccount(ctx, orgID, job); err != nil {
		return fmt.Errorf("unable to provision job service account: %w", err)
	}
	return nil
}

func orgIDForJob(ctx context.Context, h *Helpers, job *app.RunnerJob) (string, error) {
	if orgID, err := cctx.OrgIDFromContext(ctx); err == nil && orgID != "" {
		return orgID, nil
	}
	if job.RunnerID == "" {
		return "", fmt.Errorf("job %s has no org context or runner", job.ID)
	}
	var runner app.Runner
	if err := h.db.WithContext(ctx).Select("id", "org_id").Where(app.Runner{ID: job.RunnerID}).Take(&runner).Error; err != nil {
		return "", fmt.Errorf("load runner for job credentials: %w", err)
	}
	if runner.OrgID == "" {
		return "", fmt.Errorf("runner %s has no org", job.RunnerID)
	}
	return runner.OrgID, nil
}
