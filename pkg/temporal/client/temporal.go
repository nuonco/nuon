package temporal

import (
	"fmt"
	"io"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/uber-go/tally/v4"
	tclient "go.temporal.io/sdk/client"
	converter "go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

const (
	defaultNamespace string = "default"
)

type ContextKey struct{}

type temporal struct {
	v *validator.Validate

	Addr        string `validate:"required"`
	Namespace   string
	Logger      *zap.Logger `validate:"required"`
	LazyLoad    bool
	Converter   converter.DataConverter
	TallyCloser io.Closer
	tallyScope  tally.Scope
	propagators []workflow.ContextPropagator

	clientOnce sync.Once
	clientErr  error
	tclient.Client
	sync.RWMutex
}

var _ Client = (*temporal)(nil)

func New(v *validator.Validate, opts ...temporalOption) (*temporal, error) {
	logger, _ := zap.NewProduction(zap.WithCaller(false))

	tmp := &temporal{
		v:           v,
		Logger:      logger,
		Namespace:   defaultNamespace,
		propagators: make([]workflow.ContextPropagator, 0),
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

func WithContextPropagator(propagator workflow.ContextPropagator) temporalOption {
	return func(t *temporal) error {
		t.propagators = append(t.propagators, propagator)
		return nil
	}
}

func WithContextPropagators(propagators []workflow.ContextPropagator) temporalOption {
	return func(t *temporal) error {
		t.propagators = propagators
		return nil
	}
}

func WithMetricsWriter(mw metrics.Writer) temporalOption {
	return func(t *temporal) error {
		// TODO(sdboyer) it would be great if this errored instead, but there's a bunch of downstream refactors that would entail
		if mw != nil {
			t.tallyScope, t.TallyCloser = metrics.NewTallyScope(mw)
		}
		return nil
	}
}
