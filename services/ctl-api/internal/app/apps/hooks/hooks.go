package hooks

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/metrics"
	temporalclient "github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.uber.org/zap"
)

type hooks struct {
	l             *zap.Logger
	client        temporalclient.Client
	metricsWriter metrics.Writer
}

func New(v *validator.Validate, l *zap.Logger, client temporalclient.Client) app.AppHooks {
	return &hooks{
		client: client,
	}
}
