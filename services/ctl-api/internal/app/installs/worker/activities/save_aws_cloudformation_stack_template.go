package activities

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type SaveAWSCloudFormationStackVersionTemplateRequest struct {
	ID       string `validate:"required"`
	Template []byte `validate:"required"`
	Checksum string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) SaveAWSCloudFormationStackVersionTemplate(ctx context.Context, req *SaveAWSCloudFormationStackVersionTemplateRequest) error {
	obj := &app.InstallAWSCloudFormationStackVersion{
		ID: req.ID,
	}

	res := a.db.WithContext(ctx).
		Model(&obj).Updates(app.InstallAWSCloudFormationStackVersion{
		Contents: req.Template,
	})

	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update stack version")
	}
	if res.RowsAffected != 1 {
		return errors.Wrap(gorm.ErrRecordNotFound, "cloudformation stack not found")
	}

	return nil
}
