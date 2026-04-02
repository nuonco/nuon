package helpers

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	V           *validator.Validate
	Cfg         *internal.Config
	DB          *gorm.DB `name:"psql"`
	EVClient    eventloop.Client
	AcctClient  *account.Client
	QueueClient *queueclient.Client
}

type Helpers struct {
	v           *validator.Validate
	cfg         *internal.Config
	db          *gorm.DB
	evClient    eventloop.Client
	acctClient  *account.Client
	queueClient *queueclient.Client
}

func New(params Params) *Helpers {
	return &Helpers{
		v:           params.V,
		cfg:         params.Cfg,
		db:          params.DB,
		evClient:    params.EVClient,
		acctClient:  params.AcctClient,
		queueClient: params.QueueClient,
	}
}
