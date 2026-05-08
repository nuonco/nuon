package account

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/analytics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
)

const (
	accountCacheSize = 4096
	accountCacheTTL  = 30 * time.Second
)

type Params struct {
	fx.In

	Cfg             *internal.Config
	AnalyticsClient analytics.Writer
	DB              *gorm.DB `name:"psql"`
	V               *validator.Validate
	AuthzClient     *authz.Client
	EvClient        eventloop.Client
}

type Client struct {
	cfg             *internal.Config
	db              *gorm.DB
	v               *validator.Validate
	analyticsClient analytics.Writer
	authzClient     *authz.Client
	evClient        eventloop.Client
	accountCache    *expirable.LRU[string, *app.Account]
}

func New(params Params) *Client {
	return &Client{
		v:               params.V,
		cfg:             params.Cfg,
		db:              params.DB,
		analyticsClient: params.AnalyticsClient,
		authzClient:     params.AuthzClient,
		evClient:        params.EvClient,
		accountCache:    expirable.NewLRU[string, *app.Account](accountCacheSize, nil, accountCacheTTL),
	}
}

func (c *Client) InvalidateAccount(id string) {
	c.accountCache.Remove(id)
}
