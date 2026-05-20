package helpers

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	Cfg         *internal.Config
	DB          *gorm.DB `name:"psql"`
	QueueClient *queueclient.Client
}

type Helpers struct {
	cfg         *internal.Config
	db          *gorm.DB
	queueClient *queueclient.Client
}

func New(params Params) *Helpers {
	return &Helpers{
		cfg:         params.Cfg,
		db:          params.DB,
		queueClient: params.QueueClient,
	}
}
