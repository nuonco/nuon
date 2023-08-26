package hooks

import (
	"github.com/go-playground/validator/v10"
	temporalclient "github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"go.uber.org/zap"
)

const (
	defaultNamespace string = "releases"
)

type Hooks struct {
	l      *zap.Logger
	client temporalclient.Client
	cfg    *internal.Config
}

func New(v *validator.Validate, l *zap.Logger, client temporalclient.Client, cfg *internal.Config) *Hooks {
	return &Hooks{
		client: client,
		l:      l,
		cfg:    cfg,
	}
}
