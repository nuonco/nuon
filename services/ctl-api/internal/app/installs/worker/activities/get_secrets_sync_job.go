package activities

import (
	"context"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type GetSecretsSyncJobRequest struct {
	InstallID string `validate:"required"`
}

// @temporal-gen activity
// @by-id InstallID
func (a *Activities) GetSecretsSyncJob(ctx context.Context, req GetSecretsSyncJobRequest) (*app.RunnerJob, error) {
	install, err := a.getInstall(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	job := app.RunnerJob{}
	res := a.db.WithContext(ctx).
		Where(app.RunnerJob{
			Type:     app.RunnerJobTypeSandboxSyncSecrets,
			RunnerID: install.RunnerID,
		}).
		Order("created_at desc").
		Limit(1).
		First(&job)

	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get runner job")
	}

	return &job, nil
}
