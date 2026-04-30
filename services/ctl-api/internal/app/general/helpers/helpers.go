package helpers

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	QueueClient *queueclient.Client
}

type Helpers struct {
	db          *gorm.DB
	queueClient *queueclient.Client
}

func New(params Params) *Helpers {
	return &Helpers{
		db:          params.DB,
		queueClient: params.QueueClient,
	}
}
