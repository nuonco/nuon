package helpers

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/go-github/v50/github"
	"go.uber.org/fx"
	"go.uber.org/zap"
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
	L             *zap.Logger
	QueueClient   *queueclient.Client
	EmitterClient *emitterclient.Client
}

type Helpers struct {
	cfg           *internal.Config
	ghClient      *github.Client
	db            *gorm.DB
	l             *zap.Logger
	queueClient   *queueclient.Client
	emitterClient *emitterclient.Client
}

func New(params Params) *Helpers {
	return &Helpers{
		cfg:           params.Cfg,
		ghClient:      params.GhClient,
		db:            params.DB,
		l:             params.L,
		queueClient:   params.QueueClient,
		emitterClient: params.EmitterClient,
	}
}

// Logger returns the helpers logger (may be nil in tests).
func (h *Helpers) Logger() *zap.Logger {
	return h.l
}
