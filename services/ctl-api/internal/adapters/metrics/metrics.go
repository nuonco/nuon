package metrics

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"go.uber.org/zap"
)

func New(v *validator.Validate, l *zap.Logger, cfg *internal.Config) (metrics.Writer, error) {
	tags := []string{
		fmt.Sprintf("git_ref:%s", cfg.GitRef),
		fmt.Sprintf("service:%s", cfg.ServiceName),
	}

	mw, err := metrics.New(v,
		metrics.WithDisable(cfg.DisableMetrics),
		metrics.WithTags(tags...),
		metrics.WithLogger(l),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create new metrics writer: %w", err)
	}

	return mw, nil
}
