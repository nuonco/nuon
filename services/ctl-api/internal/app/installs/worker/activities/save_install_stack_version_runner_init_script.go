package activities

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type SaveInstallStackVersionRunnerInitScriptRequest struct {
	ID       string `validate:"required"`
	URL      string `validate:"required"`
	Checksum string `validate:"required"`
}

// @temporal-gen-v2 activity
func (a *Activities) SaveInstallStackVersionRunnerInitScript(ctx context.Context, req *SaveInstallStackVersionRunnerInitScriptRequest) error {
	obj := &app.InstallStackVersion{ID: req.ID}

	res := a.db.WithContext(ctx).
		Model(&obj).Updates(app.InstallStackVersion{
		RunnerInitScriptURL:      req.URL,
		RunnerInitScriptChecksum: req.Checksum,
	})

	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update stack version runner init script")
	}
	if res.RowsAffected != 1 {
		return errors.Wrap(gorm.ErrRecordNotFound, "install stack version not found")
	}

	return nil
}
