package activities

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type SaveInstallStackVersionTerraformRequest struct {
	ID       string `validate:"required"`
	Template []byte `validate:"required"`
	Checksum string `validate:"required"`
}

// SaveInstallStackVersionTerraform writes the rendered Terraform tfvars
// envelope to the install stack version. This runs alongside
// SaveInstallStackVersionTemplate (which writes CloudFormation contents) so
// both artifacts are available during the await step for the user to choose.
//
// @temporal-gen-v2 activity
func (a *Activities) SaveInstallStackVersionTerraform(ctx context.Context, req *SaveInstallStackVersionTerraformRequest) error {
	obj := &app.InstallStackVersion{ID: req.ID}

	res := a.db.WithContext(ctx).
		Model(&obj).Updates(app.InstallStackVersion{
		TerraformContents: req.Template,
		TerraformChecksum: req.Checksum,
	})

	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update stack version terraform contents")
	}
	if res.RowsAffected != 1 {
		return errors.Wrap(gorm.ErrRecordNotFound, "install stack version not found")
	}

	return nil
}
