package helpers

import (
	"context"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

type Helpers struct {
	cfg              *internal.Config
	imageAuth        *sharedactivities.Activities
	issuer           *oidcissuer.Issuer
	v                *validator.Validate
	db               *gorm.DB
	enqueueOrgSignal func(context.Context, orgshelpers.EnqueueOrgSignalParams) error
	queueClient      *queueclient.Client
}

type Params struct {
	fx.In
	Cfg         *internal.Config
	V           *validator.Validate
	DB          *gorm.DB `name:"psql"`
	SharedActs  *sharedactivities.Activities
	OrgsHelpers *orgshelpers.Helpers
	QueueClient *queueclient.Client
}

func New(params Params) (*Helpers, error) {
	h := &Helpers{cfg: params.Cfg, imageAuth: params.SharedActs, v: params.V, db: params.DB, enqueueOrgSignal: params.OrgsHelpers.EnqueueOrgSignal, queueClient: params.QueueClient}
	issuer, _, err := cloudconnections.IssuerFromConfig(params.Cfg)
	if err != nil {
		return nil, err
	}
	h.issuer = issuer
	return h, nil
}
