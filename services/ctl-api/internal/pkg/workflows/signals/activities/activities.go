package activities

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	temporalclient "github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type Params struct {
	fx.In

	EVClient eventloop.Client
	Cfg      *internal.Config
	Client   temporalclient.Client
	L        *zap.Logger
	DB       *gorm.DB `name:"psql"`
}

type Activities struct {
	evClient eventloop.Client
	client   temporalclient.Client
	cfg      *internal.Config
	l        *zap.Logger
	db       *gorm.DB
}

func New(params Params) *Activities {
	return &Activities{
		evClient: params.EVClient,
		client:   params.Client,
		cfg:      params.Cfg,
		l:        params.L,
		db:       params.DB,
	}
}
