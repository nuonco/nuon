package psql

import (
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/routing"
)

func replicaACL(db *gorm.DB) *routing.TableACL {
	return routing.NewACLBuilder(db).
		Allow(
			&app.Org{},
			&app.App{},
			&app.Install{},
			&app.Account{},
		).
		Build()
}
