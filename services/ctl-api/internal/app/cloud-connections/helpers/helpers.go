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

type imageAuthActivities interface {
	GetECRAccessToken(context.Context, *sharedactivities.GetECRAccessTokenRequest) (*sharedactivities.ECRAccessToken, error)
	GetGARAccessToken(context.Context, *sharedactivities.GetGARAccessTokenRequest) (*sharedactivities.GARAccessToken, error)
}

type Helpers struct {
	cfg              *internal.Config
	imageAuth        imageAuthActivities
	issuer           *oidcissuer.Issuer
	db               *gorm.DB
	enqueueOrgSignal func(context.Context, orgshelpers.EnqueueOrgSignalParams) error
	queueClient      *queueclient.Client
}

type Params struct {
	fx.In
	Cfg         *internal.Config
	V           *validator.Validate
	DB          *gorm.DB `name:"psql"`
	OrgsHelpers *orgshelpers.Helpers
	QueueClient *queueclient.Client
}

func New(params Params) (*Helpers, error) {
	imageAuth, err := sharedactivities.New(sharedactivities.Params{Cfg: params.Cfg, V: params.V})
	if err != nil {
		return nil, err
	}
	h := &Helpers{cfg: params.Cfg, imageAuth: imageAuth, db: params.DB, enqueueOrgSignal: params.OrgsHelpers.EnqueueOrgSignal, queueClient: params.QueueClient}
	issuer, err := cloudconnections.IssuerFromConfig(params.Cfg)
	if err != nil {
		return nil, err
	}
	h.issuer = issuer
	return h, nil
}
