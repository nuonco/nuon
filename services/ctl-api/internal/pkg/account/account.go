package account

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/analytics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
)

type Params struct {
	fx.In

	Cfg             *internal.Config
	AnalyticsClient analytics.Writer
	DB              *gorm.DB `name:"psql"`
	V               *validator.Validate
	AuthzClient     *authz.Client
}

type Client struct {
	cfg             *internal.Config
	db              *gorm.DB
	v               *validator.Validate
	analyticsClient analytics.Writer
	authzClient     *authz.Client
}

func New(params Params) *Client {
	if params.AuthzClient == nil {
		params.AuthzClient = authz.New(authz.Params{DB: params.DB})
	}
	return &Client{
		v:               params.V,
		cfg:             params.Cfg,
		db:              params.DB,
		analyticsClient: params.AnalyticsClient,
		authzClient:     params.AuthzClient,
	}
}

func (c *Client) withDB(db *gorm.DB) *Client {
	client := *c
	client.db = db
	client.authzClient = authz.New(authz.Params{DB: db})
	return &client
}
