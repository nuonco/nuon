package helpers

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/go-github/v50/github"
	"go.uber.org/fx"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
)

type Params struct {
	fx.In

	V             *validator.Validate
	Cfg           *internal.Config
	GhClient      *github.Client
	DB            *gorm.DB `name:"psql"`
	QueueClient   *queueclient.Client
	EmitterClient *emitterclient.Client
}

type Helpers struct {
	cfg           *internal.Config
	ghClient      *github.Client
	db            *gorm.DB
	queueClient   *queueclient.Client
	emitterClient *emitterclient.Client
}

func New(params Params) *Helpers {
	return &Helpers{
		cfg:           params.Cfg,
		ghClient:      params.GhClient,
		db:            params.DB,
		queueClient:   params.QueueClient,
		emitterClient: params.EmitterClient,
	}
}
