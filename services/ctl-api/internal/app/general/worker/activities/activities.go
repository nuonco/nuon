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
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/autolink"
	slackclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/client"
)

type Activities struct {
	cfg            *internal.Config
	db             *gorm.DB
	chDB           *gorm.DB
	appsHelpers    *appshelpers.Helpers
	runnersHelpers *runnershelpers.Helpers
	mw             metrics.Writer
	logger         *temporalzap.Logger
	l              *zap.Logger
	tClient        temporalclient.Client
	slackClient    *slackclient.Client
	autoLinkHelper *autolink.Helper
	blobSvc        blobstore.Service
}

type Params struct {
	fx.In

	Cfg            *internal.Config
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	AppsHelpers    *appshelpers.Helpers
	RunnersHelpers *runnershelpers.Helpers
	MW             metrics.Writer
	TemporalClient temporalclient.Client
	SlackClient    *slackclient.Client
	AutoLinkHelper *autolink.Helper
	BlobSvc        blobstore.Service
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
		runnersHelpers: params.RunnersHelpers,
		mw:             params.MW,
		logger:         tlogger,
		l:              logger,
		tClient:        params.TemporalClient,
		slackClient:    params.SlackClient,
		autoLinkHelper: params.AutoLinkHelper,
		blobSvc:        params.BlobSvc,
	}, nil
}
