package notifications

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/loops"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
)

type Notifications struct {
	Loops loops.Client     `validate:"required"`
	Cfg   *internal.Config `validate:"required"`
	L     *zap.Logger      `validate:"required"`

	MetricsWriter metrics.Writer `validate:"required"`
}

func New(v *validator.Validate,
	l *zap.Logger,
	cfg *internal.Config,
	loopsclient loops.Client,
	metricsWriter metrics.Writer,
) (*Notifications, error) {
	not := &Notifications{
		Cfg:           cfg,
		Loops:         loopsclient,
		L:             l,
		MetricsWriter: metricsWriter,
	}
	if err := v.Struct(not); err != nil {
		return nil, fmt.Errorf("unable to validate: %w", err)
	}

	return not, nil
}
