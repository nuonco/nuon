package activities

import (
	"context"

	"gorm.io/gorm"

	pkggenerics "github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

type UpdateInstallStackOutputs struct {
	InstallStackID           string            `validate:"required"`
	InstallStackVersionRunID string            `validate:"required"`
	Data                     map[string]string `validate:"required"`
}

// @temporal-gen activity
func (a *Activities) UpdateInstallStackOutputs(ctx context.Context, req UpdateInstallStackOutputs) error {
	outputs := app.InstallStackOutputs{}
	res := a.db.WithContext(ctx).
		Model(&outputs).
		Where(app.InstallStackOutputs{
			InstallStackID: req.InstallStackID,
		}).
		Updates(app.InstallStackOutputs{
			Data:                     generics.ToHstore(req.Data),
			InstallStackVersionRunID: pkggenerics.NewNullString(req.InstallStackVersionRunID),
		})
	if res.Error != nil {
		return generics.TemporalGormError(res.Error)
	}
	if res.RowsAffected < 1 {
		return generics.TemporalGormError(gorm.ErrRecordNotFound)
	}

	return nil
}
