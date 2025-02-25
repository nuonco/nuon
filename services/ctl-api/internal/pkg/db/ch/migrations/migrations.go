package migrations

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"go.uber.org/fx"
)

func All() []migrations.Migration {
	return []migrations.Migration{}
}

type Params struct {
	fx.In
}

type Migrations struct{}

func New(params Params) *Migrations {
	return &Migrations{}
}
