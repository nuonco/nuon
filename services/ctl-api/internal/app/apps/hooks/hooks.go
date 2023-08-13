package hooks

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/metrics"
	temporalclient "github.com/powertoolsdev/mono/pkg/temporal/client"
	"go.uber.org/zap"
)

type Hooks struct {
	l             *zap.Logger
	client        temporalclient.Client
	metricsWriter metrics.Writer
}

func New(v *validator.Validate, l *zap.Logger, client temporalclient.Client) *Hooks {
	return &Hooks{
		client: client,
		l:      l,
	}
}
