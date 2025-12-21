package activities

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	V           *validator.Validate
	DB          *gorm.DB `name:"psql"`
	QueueClient *queueclient.Client
	L           *zap.Logger
}

type Activities struct {
	v           *validator.Validate
	db          *gorm.DB
	queueClient *queueclient.Client
	l           *zap.Logger
}

func New(params Params) *Activities {
	return &Activities{
		v:           params.V,
		db:          params.DB,
		queueClient: params.QueueClient,
		l:           params.L,
	}
}
