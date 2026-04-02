package helpers

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
)

type Params struct {
	fx.In

	V              *validator.Validate
	Cfg            *internal.Config
	DB             *gorm.DB `name:"psql"`
	EVClient       eventloop.Client
	AcctClient     *account.Client
	QueueClient    *queueclient.Client
	EmitterClient  *emitterclient.Client
	FeaturesClient *features.Features
}

type Helpers struct {
	v              *validator.Validate
	cfg            *internal.Config
	db             *gorm.DB
	evClient       eventloop.Client
	acctClient     *account.Client
	queueClient    *queueclient.Client
	emitterClient  *emitterclient.Client
	featuresClient *features.Features
}

func New(params Params) *Helpers {
	return &Helpers{
		v:              params.V,
		cfg:            params.Cfg,
		db:             params.DB,
		evClient:       params.EVClient,
		acctClient:     params.AcctClient,
		queueClient:    params.QueueClient,
		emitterClient:  params.EmitterClient,
		featuresClient: params.FeaturesClient,
	}
}
