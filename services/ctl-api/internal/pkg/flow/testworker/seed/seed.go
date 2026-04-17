package seed

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Params struct {
	fx.In

	L           *zap.Logger
	DB          *gorm.DB `name:"psql"`
	AcctClient  *account.Client
	QueueClient *queueclient.Client
}

type Seeder struct {
	db          *gorm.DB
	l           *zap.Logger
	acctHelpers *account.Client
	queueClient *queueclient.Client
}

func New(params Params) *Seeder {
	return &Seeder{
		db:          params.DB,
		l:           params.L,
		acctHelpers: params.AcctClient,
		queueClient: params.QueueClient,
	}
}
