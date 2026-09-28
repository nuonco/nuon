package helpers

import (
	"context"

	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type Helpers struct {
	issuer           *oidcissuer.Issuer
	db               *gorm.DB
	enqueueOrgSignal func(context.Context, orgshelpers.EnqueueOrgSignalParams) error
	queueClient      *queueclient.Client
}

type Params struct {
	fx.In
	Cfg         *internal.Config
	DB          *gorm.DB `name:"psql"`
	OrgsHelpers *orgshelpers.Helpers
	QueueClient *queueclient.Client
}

func New(params Params) (*Helpers, error) {
	h := &Helpers{db: params.DB, enqueueOrgSignal: params.OrgsHelpers.EnqueueOrgSignal, queueClient: params.QueueClient}
	issuer, err := cloudconnections.IssuerFromConfig(params.Cfg)
	if err != nil {
		return nil, err
	}
	h.issuer = issuer
	return h, nil
}
