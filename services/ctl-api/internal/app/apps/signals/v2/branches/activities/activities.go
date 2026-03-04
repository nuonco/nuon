package activities

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type Params struct {
	fx.In

	V           *validator.Validate
	Helpers     *helpers.Helpers
	DB          *gorm.DB `name:"psql"`
	BlobService blobstore.Service
}

type Activities struct {
	v       *validator.Validate
	db      *gorm.DB
	helpers *helpers.Helpers
	blobSvc blobstore.Service
}

func New(params Params) (*Activities, error) {
	return &Activities{
		v:       params.V,
		db:      params.DB,
		helpers: params.Helpers,
		blobSvc: params.BlobService,
	}, nil
}
