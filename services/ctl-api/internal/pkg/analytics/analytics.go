package analytics

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/analytics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"go.uber.org/zap"
)

func New(v *validator.Validate, l *zap.Logger, cfg *internal.Config) (analytics.Client, error) {
	client, err := analytics.New(cfg.SegmentWriteKey)
	if err != nil {
		return nil, fmt.Errorf("unable to create new analytics client: %w", err)
	}

	return client, nil
}
