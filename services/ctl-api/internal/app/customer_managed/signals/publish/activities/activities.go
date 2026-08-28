package activities

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed/transport"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type Params struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	V           *validator.Validate
	Store       transport.Store
	Config      *internal.Config
	AppsHelpers *appshelpers.Helpers
	BlobService blobstore.Service
}

type Activities struct {
	db          *gorm.DB
	v           *validator.Validate
	store       transport.Store
	cfg         *internal.Config
	appsHelpers *appshelpers.Helpers
	blobSvc     blobstore.Service
}

func New(params Params) *Activities {
	return &Activities{db: params.DB, v: params.V, store: params.Store, cfg: params.Config, appsHelpers: params.AppsHelpers, blobSvc: params.BlobService}
}
