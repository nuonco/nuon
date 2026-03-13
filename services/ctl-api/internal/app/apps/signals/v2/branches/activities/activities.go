package activities

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type Params struct {
	fx.In

	V           *validator.Validate
	Helpers     *helpers.Helpers
	DB          *gorm.DB `name:"psql"`
	BlobService blobstore.Service
	AcctClient  *account.Client
	AuthzClient *authz.Client
	Cfg         *internal.Config
	L           *zap.Logger
}

type Activities struct {
	v           *validator.Validate
	db          *gorm.DB
	helpers     *helpers.Helpers
	blobSvc     blobstore.Service
	acctClient  *account.Client
	authzClient *authz.Client
	cfg         *internal.Config
	l           *zap.Logger
}

func New(params Params) (*Activities, error) {
	return &Activities{
		v:           params.V,
		db:          params.DB,
		helpers:     params.Helpers,
		blobSvc:     params.BlobService,
		acctClient:  params.AcctClient,
		authzClient: params.AuthzClient,
		cfg:         params.Cfg,
		l:           params.L,
	}, nil
}
