package activities

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
)

type Params struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	CHDB        *gorm.DB `name:"ch"`
	AppsHelpers *appshelpers.Helpers
	TClient     temporalclient.Client
	Cfg         *internal.Config
	L           *zap.Logger `optional:"true"`
}

type Activities struct {
	db          *gorm.DB
	chDB        *gorm.DB
	appsHelpers *appshelpers.Helpers
	tClient     temporalclient.Client
	cfg         *internal.Config
	l           *zap.Logger
}

func New(params Params) *Activities {
	l := params.L
	if l == nil {
		l = zap.NewNop()
	}
	return &Activities{
		db:          params.DB,
		chDB:        params.CHDB,
		appsHelpers: params.AppsHelpers,
		tClient:     params.TClient,
		cfg:         params.Cfg,
		l:           l,
	}
}
