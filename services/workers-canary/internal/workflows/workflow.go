package workflows

import (
	"time"

	"go.uber.org/zap"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/temporal/metrics"
	workers "github.com/powertoolsdev/mono/services/workers-canary/internal"
	"github.com/powertoolsdev/mono/services/workers-canary/internal/activities"
)

const (
	defaultStartActivityTimeout time.Duration = time.Second * 5
	defaultPollActivityTimeout  time.Duration = time.Minute * 30
	defaultMaxActivityRetries                 = 5
	defaultRegion                             = "us-west-2"
)

type wkflow struct {
	cfg           workers.Config
	acts          *activities.Activities
	l             *zap.Logger
	metricsWriter metrics.Writer
}

func New(v *validator.Validate, cfg workers.Config, metricsWriter metrics.Writer) (*wkflow, error) {
	return &wkflow{
		cfg:           cfg,
		metricsWriter: metricsWriter,
		l:             zap.L(),
	}, nil
}
