package client

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/temporal/temporalzap"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

type temporal struct {
	v *validator.Validate

	Addr      string      `validate:"required"`
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
		HostPort:  tmp.Addr,
		Namespace: tmp.Namespace,
		Logger:    temporalzap.NewLogger(logger),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to dial temporal: %w", err)
	}

	return tc, nil
}

type temporalOption func(*temporal) error

func WithAddr(addr string) temporalOption {
	return func(t *temporal) error {
		t.Addr = addr
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
