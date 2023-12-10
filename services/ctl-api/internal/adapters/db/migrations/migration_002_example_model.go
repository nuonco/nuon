package migrations

import (
	"context"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (a *Migrations) migration002ExampleModel(ctx context.Context) error {
	// sample code using gorm
	var installs []*app.Install
	res := a.db.WithContext(ctx).
		Find(&installs)

	if res.Error != nil {
		return res.Error
	}

	for range installs {
		a.l.Info("example migration - gorm update")
	}

	return nil
}
