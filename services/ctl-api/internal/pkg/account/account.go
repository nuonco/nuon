package account

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/analytics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

type Params struct {
	fx.In

	Cfg             *internal.Config
	AnalyticsClient analytics.Writer
	DB              *gorm.DB `name:"psql"`
	V               *validator.Validate
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
