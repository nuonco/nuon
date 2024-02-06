package worker

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	"go.uber.org/zap"
)

type Workflows struct {
	cfg           *internal.Config
	v             *validator.Validate
	acts          activities.Activities
	l             *zap.Logger
	metricsWriter metrics.Writer
}

func NewWorkflows(v *validator.Validate,
	cfg *internal.Config,
	l *zap.Logger,
	metricsWriter metrics.Writer,
) *Workflows {
	return &Workflows{
		cfg:           cfg,
		v:             v,
		l:             l,
		metricsWriter: metricsWriter,
		//  NOTE: this field is only used to be able to fetch activity methods
		acts: activities.Activities{},
	}
}
