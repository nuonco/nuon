package activities

import (
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	slackclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/client"
)

type Activities struct {
	cfg         *internal.Config
	db          *gorm.DB
	chDB        *gorm.DB
	appsHelpers *appshelpers.Helpers
	evClient    eventloop.Client
	mw          metrics.Writer
	logger      *temporalzap.Logger
	tClient     temporalclient.Client
	slackClient *slackclient.Client
}

type Params struct {
	fx.In

	Cfg            *internal.Config
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	AppsHelpers    *appshelpers.Helpers
	EvClient       eventloop.Client
	MW             metrics.Writer
	TemporalClient temporalclient.Client
	SlackClient    *slackclient.Client
}

func New(params Params) (*Activities, error) {
	logger, err := zap.NewProduction()
	tlogger := temporalzap.NewLogger(logger)
	if err != nil {
		return nil, fmt.Errorf("unable to create temporal logger: %w", err)
	}
	return &Activities{
		cfg:         params.Cfg,
		db:          params.DB,
		chDB:        params.CHDB,
		appsHelpers: params.AppsHelpers,
		evClient:    params.EvClient,
		mw:          params.MW,
		logger:      tlogger,
		tClient:     params.TemporalClient,
		slackClient: params.SlackClient,
	}, nil
}
