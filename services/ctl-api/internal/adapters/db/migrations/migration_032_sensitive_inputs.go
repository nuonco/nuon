package migrations

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (a *Migrations) migration033SensitiveInputs(ctx context.Context) error {
	res := a.db.Unscoped().WithContext(ctx).
		Model(&app.AppInput{}).
		Updates(app.AppInput{
			Sensitive: false,
		})
	if res.Error != nil {
		return res.Error
	}

	return nil
}
