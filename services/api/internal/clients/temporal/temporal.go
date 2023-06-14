package temporal

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/temporal/temporalzap"
	"github.com/powertoolsdev/mono/services/api/internal"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

type temporal struct {
	HostPort  string
	Namespace string
	Logger    *zap.Logger
}

func New(opts ...temporalOption) (tclient.Client, error) {
	logger, _ := zap.NewProduction(zap.WithCaller(false))
	tmp := &temporal{
		Logger: logger,
	}

	for idx, opt := range opts {
		if err := opt(tmp); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	validate := validator.New()
	if err := validate.Struct(tmp); err != nil {
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
		t.HostPort = cfg.TemporalHost
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

func WithNamespace(namespace string) temporalOption {
	return func(t *temporal) error {
		t.Namespace = namespace
		return nil
	}
}
