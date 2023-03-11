package temporal

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/common/temporalzap"
	"github.com/powertoolsdev/mono/services/nuonctl/internal"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

type temporal struct {
	v *validator.Validate

	HostPort  string      `validate:"required"`
	Namespace string      `validate:"required"`
	Logger    *zap.Logger `validate:"required"`
}

func New(v *validator.Validate, opts ...temporalOption) (tclient.Client, error) {
	logger, _ := zap.NewProduction(zap.WithCaller(false))
	tmp := &temporal{
		v:      v,
		Logger: logger,
	}

	for idx, opt := range opts {
		if err := opt(tmp); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	if err := v.Struct(tmp); err != nil {
		return nil, fmt.Errorf("unable to validate temporal: %w", err)
	}

	tc, err := tclient.Dial(tclient.Options{
		HostPort:  tmp.HostPort,
		Namespace: tmp.Namespace,
		Logger:    temporalzap.NewLogger(logger),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to dial temporal: %w", err)
	}

	return tc, nil
}

type temporalOption func(*temporal) error

func WithConfig(cfg *internal.Config) temporalOption {
	return func(t *temporal) error {
		if cfg.Env == "stage" {
			t.HostPort = cfg.StageTemporalHost
		} else {
			t.HostPort = cfg.DevTemporalHost
		}

		t.Namespace = cfg.TemporalNamespace
		return nil
	}
}

func WithLogger(log *zap.Logger) temporalOption {
	return func(t *temporal) error {
		t.Logger = log
		return nil
	}
}
