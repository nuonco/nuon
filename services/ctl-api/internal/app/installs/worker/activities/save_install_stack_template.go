package activities

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type SaveInstallStackVersionTemplateRequest struct {
	ID                 string            `validate:"required"`
	Template           []byte            `validate:"required"`
	Checksum           string            `validate:"required"`
	AWSStackParameters map[string]string `temporaljson:"aws_stack_parameters,omitempty"`
}

// @temporal-gen-v2 activity
func (a *Activities) SaveInstallStackVersionTemplate(ctx context.Context, req *SaveInstallStackVersionTemplateRequest) error {
	obj := &app.InstallStackVersion{
		ID: req.ID,
	}

	res := a.db.WithContext(ctx).
		Model(&obj).Select("Contents", "AWSStackParameters").Updates(app.InstallStackVersion{
		Contents:           req.Template,
		AWSStackParameters: req.AWSStackParameters,
	})

	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update stack version")
	}
	if res.RowsAffected != 1 {
		return errors.Wrap(gorm.ErrRecordNotFound, "cloudformation stack not found")
	}

	return nil
}
