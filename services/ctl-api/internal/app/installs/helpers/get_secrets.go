package helpers

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (h *Helpers) getSecrets(ctx context.Context, installID, runnerID string) (*state.SecretsState, error) {
	runnerJob, err := h.getSecretsSyncRunnerJob(ctx, installID, runnerID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Wrap(err, "unable to get secrets")
		}
		return ToSecretsState(nil), nil
	}
	return ToSecretsState(runnerJob.ParsedOutputs), nil
}

func (h *Helpers) getSecretsSyncRunnerJob(ctx context.Context, installID, runnerID string) (*app.RunnerJob, error) {
	job := app.RunnerJob{}
	res := h.db.WithContext(ctx).
		Where(app.RunnerJob{
			Type:     app.RunnerJobTypeSandboxSyncSecrets,
			RunnerID: runnerID,
		}).
		Order("created_at desc").
		Limit(1).
		First(&job)

	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get runner job")
	}

	return &job, nil
}
