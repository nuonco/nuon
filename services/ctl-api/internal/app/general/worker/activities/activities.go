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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/autolink"
	slackclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/client"
)

type Activities struct {
	cfg            *internal.Config
	db             *gorm.DB
	chDB           *gorm.DB
	appsHelpers    *appshelpers.Helpers
	mw             metrics.Writer
	logger         *temporalzap.Logger
	tClient        temporalclient.Client
	slackClient    *slackclient.Client
	autoLinkHelper *autolink.Helper
}

type Params struct {
	fx.In

	Cfg            *internal.Config
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	AppsHelpers    *appshelpers.Helpers
	MW             metrics.Writer
	TemporalClient temporalclient.Client
	SlackClient    *slackclient.Client
	AutoLinkHelper *autolink.Helper
}

func New(params Params) (*Activities, error) {
	logger, err := zap.NewProduction()
	tlogger := temporalzap.NewLogger(logger)
	if err != nil {
		return nil, fmt.Errorf("unable to create temporal logger: %w", err)
	}
	return &Activities{
		cfg:            params.Cfg,
		db:             params.DB,
		chDB:           params.CHDB,
		appsHelpers:    params.AppsHelpers,
		mw:             params.MW,
		logger:         tlogger,
		tClient:        params.TemporalClient,
		slackClient:    params.SlackClient,
		autoLinkHelper: params.AutoLinkHelper,
	}, nil
}
