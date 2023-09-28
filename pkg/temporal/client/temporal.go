package temporal

import (
	"fmt"
	"sync"

	"github.com/go-playground/validator/v10"
	tclient "go.temporal.io/sdk/client"
	converter "go.temporal.io/sdk/converter"
	"go.uber.org/zap"
)

const (
	defaultNamespace string = "default"
)

type ContextKey struct{}

type temporal struct {
	v *validator.Validate

	Addr      string      `validate:"required"`
	Namespace string      `validate:"required"`
	Logger    *zap.Logger `validate:"required"`
	LazyLoad  bool
	Converter converter.DataConverter

	tclient.Client
	sync.RWMutex
}

var _ Client = (*temporal)(nil)

func New(v *validator.Validate, opts ...temporalOption) (*temporal, error) {
	logger, _ := zap.NewProduction(zap.WithCaller(false))
	tmp := &temporal{
		v:         v,
		Logger:    logger,
		Namespace: defaultNamespace,
	}

	for idx, opt := range opts {
		if err := opt(tmp); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	if err := v.Struct(tmp); err != nil {
		return nil, fmt.Errorf("unable to validate temporal: %w", err)
	}

	if !tmp.LazyLoad {
		if _, err := tmp.getClient(); err != nil {
			return nil, fmt.Errorf("unable to set temporal client: %w", err)
		}
	}

	return tmp, nil
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

func WithLazyLoad(lazyLoad bool) temporalOption {
	return func(t *temporal) error {
		t.LazyLoad = lazyLoad
		return nil
	}
}

func WithDataConverter(dataConverter converter.DataConverter) temporalOption {
	return func(t *temporal) error {
		t.Converter = dataConverter
		return nil
	}
}
