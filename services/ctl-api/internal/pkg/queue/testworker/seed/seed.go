package seed

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
)

type Params struct {
	fx.In

	L          *zap.Logger
	DB         *gorm.DB `name:"psql"`
	AcctClient *account.Client
}

type Seeder struct {
	db          *gorm.DB
	l           *zap.Logger
	acctHelpers *account.Client
}

func New(params Params) *Seeder {
	return &Seeder{
		db:          params.DB,
		l:           params.L,
		acctHelpers: params.AcctClient,
	}
}
