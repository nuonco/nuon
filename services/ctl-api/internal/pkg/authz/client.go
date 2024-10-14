package authz

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/analytics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type Params struct {
	fx.In

	Cfg             *internal.Config
	DB              *gorm.DB `name:"psql"`
	V               *validator.Validate
	EvClient        eventloop.Client
	AnalyticsClient analytics.Writer
}

type Client struct {
	cfg             *internal.Config
	db              *gorm.DB
	v               *validator.Validate
	evClient        eventloop.Client
	analyticsClient analytics.Writer
}

func New(params Params) *Client {
	return &Client{
		v:               params.V,
		cfg:             params.Cfg,
		db:              params.DB,
		evClient:        params.EvClient,
		analyticsClient: params.AnalyticsClient,
	}
}
