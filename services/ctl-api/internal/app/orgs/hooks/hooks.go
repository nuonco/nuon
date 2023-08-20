package hooks

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/metrics"
	temporalclient "github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"go.uber.org/zap"
)

const (
	defaultNamespace string = "orgs"
)

type Hooks struct {
	l             *zap.Logger
	client        temporalclient.Client
	metricsWriter metrics.Writer
	cfg           *internal.Config
}

func New(v *validator.Validate, l *zap.Logger, client temporalclient.Client, metricsWriter metrics.Writer, cfg *internal.Config) *Hooks {
	return &Hooks{
		client:        client,
		l:             l,
		metricsWriter: metricsWriter,
		cfg:           cfg,
	}
}
