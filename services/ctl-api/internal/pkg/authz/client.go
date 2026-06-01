package authz

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/analytics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type Params struct {
	fx.In

	Cfg             *internal.Config
	DB              *gorm.DB `name:"psql"`
	V               *validator.Validate
	AnalyticsClient analytics.Writer
}

type Client struct {
	cfg             *internal.Config
	db              *gorm.DB
	v               *validator.Validate
	analyticsClient analytics.Writer
}

func New(params Params) *Client {
	return &Client{
		v:               params.V,
		cfg:             params.Cfg,
		db:              params.DB,
		analyticsClient: params.AnalyticsClient,
	}
}
