package writer

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/workers-executors/internal/executors/waypoint/event"
	"go.uber.org/zap"
)

type logEventWriter struct {
	Logger *zap.Logger `validate:"required"`

	// internal state
	v *validator.Validate
}

var _ EventWriter = (*logEventWriter)(nil)

type logEventWriterOption func(*logEventWriter) error

// NewLog creates a new event writer that writes to the provided log
func NewLog(v *validator.Validate, opts ...logEventWriterOption) (*logEventWriter, error) {
	w := &logEventWriter{v: v}

	if v == nil {
		return nil, fmt.Errorf("error instantiating writer: validator is nil")
	}

	for _, opt := range opts {
		if err := opt(w); err != nil {
			return nil, err
		}
	}

	if err := w.v.Struct(w); err != nil {
		return nil, err
	}

	return w, nil
}

func WithLog(l *zap.Logger) logEventWriterOption {
	return func(w *logEventWriter) error {
		w.Logger = l
		return nil
	}
}

func (w *logEventWriter) Write(ev event.WaypointJobEvent) error {
	w.Logger.Debug("processed waypoint job event", zap.Any("event", ev))
	return nil
}
