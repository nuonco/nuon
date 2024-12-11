package metrics

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

const (
	defaultStatsdHostEnvVar string = "HOST_IP"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_writer.go -source=metrics.go -package=metrics
type Writer interface {
	// dogstatsd metrics
	Incr(name string, tags []string)
	Decr(name string, tags []string)
	Timing(name string, value time.Duration, tags []string)
	Gauge(name string, value float64, tags []string)
	Count(name string, value int64, tags []string)

	// datadog specific
	Event(e *statsd.Event)

	Flush()
}

type writer struct {
	v *validator.Validate

	Address string
	Disable bool
	Tags    []string
	Log     *zap.Logger `validate:"required"`

	// internal
	clientonce sync.Once
	client     dogstatsdClient
	clienterr  error
}

var _ Writer = (*writer)(nil)

// New returns a default writer, which emits metrics to statsd by default
func New(v *validator.Validate, opts ...writerOption) (*writer, error) {
	l, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("unable to get logger: %w", err)
	}
	r := &writer{
		v:    v,
		Tags: make([]string, 0),
		Log:  l,
	}
	for idx, opt := range opts {
		if err := opt(r); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	if !r.Disable && r.Address == "" {
		r.Address = fmt.Sprintf("%s:8125", os.Getenv(defaultStatsdHostEnvVar))
	}

	if err := r.v.Struct(r); err != nil {
		return nil, fmt.Errorf("unable to validate writer: %w", err)
	}

	return r, nil
}

type writerOption func(*writer) error

func WithAddress(addr string) writerOption {
	return func(w *writer) error {
		w.Address = addr
		return nil
	}
}

func WithDisable(disable bool) writerOption {
	return func(w *writer) error {
		w.Disable = disable
		return nil
	}
}

func WithTags(tags ...string) writerOption {
	return func(w *writer) error {
		for _, tag := range tags {
			if strings.Count(tag, ":") != 1 {
				return fmt.Errorf("invalid tag, must be of format `k:v`")
			}
		}

		w.Tags = append(w.Tags, tags...)
		return nil
	}
}

func WithTagMap(tags map[string]string) writerOption {
	return func(w *writer) error {
		for k, v := range tags {
			w.Tags = append(w.Tags, fmt.Sprintf("%s:%s", k, v))
		}
		return nil
	}
}

func WithLogger(log *zap.Logger) writerOption {
	return func(w *writer) error {
		w.Log = log
		return nil
	}
}
