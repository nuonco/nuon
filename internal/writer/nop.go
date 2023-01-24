package writer

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/workers-executors/internal/event"
)

type nopEventWriter struct {

	// internal state
	v *validator.Validate
}

var _ EventWriter = (*nopEventWriter)(nil)

type nopEventWriterOption func(*nopEventWriter) error

// NewNop
func NewNop(v *validator.Validate, opts ...nopEventWriterOption) (*nopEventWriter, error) {
	w := &nopEventWriter{v: v}

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

func (w *nopEventWriter) Write(ev event.WaypointJobEvent) error {
	return nil
}
