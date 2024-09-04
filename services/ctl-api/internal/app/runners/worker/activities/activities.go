package activities

import (
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/protos"
)

type Params struct {
	fx.In

	Prt           *protos.Adapter
	DB            *gorm.DB `name:"psql"`
	Helpers       *helpers.Helpers
	EVClient      eventloop.Client
	AuthzClient   *authz.Client
	AccountClient *account.Client
}

type Activities struct {
	db          *gorm.DB
	protos      *protos.Adapter
	helpers     *helpers.Helpers
	evClient    eventloop.Client
	authzClient *authz.Client
	acctClient  *account.Client
}

func New(params Params) *Activities {
	return &Activities{
		db:          params.DB,
		protos:      params.Prt,
		helpers:     params.Helpers,
		evClient:    params.EVClient,
		authzClient: params.AuthzClient,
		acctClient:  params.AccountClient,
	}
}
