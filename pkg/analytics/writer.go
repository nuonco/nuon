package analytics

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	segment "github.com/segmentio/analytics-go/v3"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/analytics/events"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=mock_writer.go -source=writer.go -package=analytics
type Writer interface {
	Identify(context.Context)
	Group(context.Context)
	Close()
	Track(context.Context, events.Event, map[string]interface{})
}

var _ Writer = (*writer)(nil)

type writer struct {
	v *validator.Validate

	SegmentKey string      `validate:"required"`
	GroupFn    GroupFn     `validate:"required"`
	IdentifyFn IdentifyFn  `validate:"required"`
	UserIDFn   UserIDFn    `validate:"required"`
	Logger     *zap.Logger `validate:"required"`
	Properties map[string]interface{}

	Disable bool

	client segment.Client
}

type optFn func(w *writer) error

func New(v *validator.Validate, opts ...optFn) (*writer, error) {
	w := &writer{
		v: v,
	}
	for idx, opt := range opts {
		if err := opt(w); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}
	if err := v.Struct(w); err != nil {
		return nil, fmt.Errorf("unable to validate struct: %w", err)
	}

	if !w.Disable {
		w.client = segment.New(w.SegmentKey)
	}

	return w, nil
}

func WithSegmentKey(key string) optFn {
	return func(w *writer) error {
		w.SegmentKey = key
		return nil
	}
}

func WithGroupFn(fn GroupFn) optFn {
	return func(w *writer) error {
		w.GroupFn = fn
		return nil
	}
}

func WithIdentifyFn(fn IdentifyFn) optFn {
	return func(w *writer) error {
		w.IdentifyFn = fn
		return nil
	}
}

func WithUserIDFn(fn UserIDFn) optFn {
	return func(w *writer) error {
		w.UserIDFn = fn
		return nil
	}
}

func WithLogger(l *zap.Logger) optFn {
	return func(w *writer) error {
		w.Logger = l
		return nil
	}
}

func WithDisable(disabled bool) optFn {
	return func(w *writer) error {
		w.Disable = disabled
		return nil
	}
}

func WithProperties(props map[string]interface{}) optFn {
	return func(w *writer) error {
		w.Properties = props
		return nil
	}
}
