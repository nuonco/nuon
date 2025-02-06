package migrations

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz"
)

type Params struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	CHDB        *gorm.DB `name:"ch"`
	L           *zap.Logger
	Cfg         *internal.Config
	AuthzClient *authz.Client
	AcctClient  *account.Client
}

type Migrations struct {
	db          *gorm.DB
	chDB        *gorm.DB
	l           *zap.Logger
	cfg         *internal.Config
	authzClient *authz.Client
	acctClient  *account.Client
}

func New(params Params) *Migrations {
	return &Migrations{
		db:          params.DB,
		chDB:        params.CHDB,
		l:           params.L,
		cfg:         params.Cfg,
		authzClient: params.AuthzClient,
		acctClient:  params.AcctClient,
	}
}
