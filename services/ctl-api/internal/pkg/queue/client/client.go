package client

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	temporalclient "github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

type Client struct {
	db      *gorm.DB
	cfg     *internal.Config
	tClient temporalclient.Client
	l       *zap.Logger
}

type Params struct {
	fx.In

	DB      *gorm.DB `name:"psql"`
	Cfg     *internal.Config
	TClient temporalclient.Client
	L       *zap.Logger
}

func New(params Params) *Client {
	return &Client{
		db:      params.DB,
		cfg:     params.Cfg,
		tClient: params.TClient,
		l:       params.L,
	}
}
