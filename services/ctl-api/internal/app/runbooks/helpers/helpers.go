package helpers

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	V           *validator.Validate
	Cfg         *internal.Config
	DB          *gorm.DB `name:"psql"`
	AppsHelpers *appshelpers.Helpers
	QueueClient *queueclient.Client
}

type Helpers struct {
	cfg         *internal.Config
	db          *gorm.DB
	appsHelpers *appshelpers.Helpers
	queueClient *queueclient.Client
}

func New(params Params) *Helpers {
	return &Helpers{
		cfg:         params.Cfg,
		db:          params.DB,
		appsHelpers: params.AppsHelpers,
		queueClient: params.QueueClient,
	}
}
