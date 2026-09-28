package activities

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
	orgshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
)

type Params struct {
	fx.In
	Cfg         *internal.Config
	DB          *gorm.DB `name:"psql"`
	L           *zap.Logger
	OrgsHelpers *orgshelpers.Helpers
}

type Activities struct {
	db               *gorm.DB
	l                *zap.Logger
	verifier         cloudconnections.Verifier
	enqueueOrgSignal func(context.Context, orgshelpers.EnqueueOrgSignalParams) error
}

func New(params Params) (*Activities, error) {
	verifier, err := cloudconnections.NewVerifierFromConfig(params.Cfg, params.L)
	if err != nil {
		return nil, err
	}
	return &Activities{db: params.DB, l: params.L, verifier: verifier, enqueueOrgSignal: params.OrgsHelpers.EnqueueOrgSignal}, nil
}
