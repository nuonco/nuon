package adminops

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	generalhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/general/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/enqueuer"
)

type Params struct {
	fx.In

	DB             *gorm.DB `name:"psql"`
	AcctClient     *account.Client
	QueueClient    *queueclient.Client
	GeneralHelpers *generalhelpers.Helpers
	Enqueuer       *enqueuer.Enqueuer
}

type Ops struct {
	db             *gorm.DB
	acctClient     *account.Client
	queueClient    *queueclient.Client
	generalHelpers *generalhelpers.Helpers
	enqueuer       *enqueuer.Enqueuer
}

func New(params Params) *Ops {
	return &Ops{
		db:             params.DB,
		acctClient:     params.AcctClient,
		queueClient:    params.QueueClient,
		generalHelpers: params.GeneralHelpers,
		enqueuer:       params.Enqueuer,
	}
}
