package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	MW          metrics.Writer
	Cfg         *internal.Config
	QueueClient *queueclient.Client
}

type Activities struct {
	db          *gorm.DB
	mw          metrics.Writer
	cfg         *internal.Config
	queueClient *queueclient.Client
}

func New(params Params) *Activities {
	return &Activities{
		db:          params.DB,
		mw:          params.MW,
		cfg:         params.Cfg,
		queueClient: params.QueueClient,
	}
}
