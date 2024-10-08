package analytics

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/analytics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"go.uber.org/zap"
)

func NewContextWriter(v *validator.Validate, l *zap.Logger, cfg *internal.Config) (*analytics.ContextWriter, error) {
	writer := analytics.NewContextWriter(cfg.SegmentWriteKey, l)

	if err := v.Struct(writer); err != nil {
		return nil, fmt.Errorf("unable to validate analytics context writer: %w", err)
	}

	return writer, nil
}

func NewTemporalWriter(v *validator.Validate, l *zap.Logger, cfg *internal.Config) (*analytics.TemporalWriter, error) {
	writer := analytics.NewTemporalWriter(cfg.SegmentWriteKey, l)

	if err := v.Struct(writer); err != nil {
		return nil, fmt.Errorf("unable to validate analytics context writer: %w", err)
	}

	return writer, nil
}
