package writer

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/workers-executors/internal/event"
)

type fileEventWriter struct {
	File *os.File `validate:"required"`

	// internal state
	v *validator.Validate
}

var _ EventWriter = (*fileEventWriter)(nil)

type fileEventWriterOption func(*fileEventWriter) error

// NewFile creates a new event writer that writes to the provided file
func NewFile(v *validator.Validate, opts ...fileEventWriterOption) (*fileEventWriter, error) {
	w := &fileEventWriter{v: v}

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

func WithFile(f *os.File) fileEventWriterOption {
	return func(few *fileEventWriter) error {
		few.File = f
		return nil
	}
}

func (w *fileEventWriter) Write(ev event.WaypointJobEvent) error {
	// convert event struct to json
	byts, err := json.Marshal(ev)
	if err != nil {
		return err
	}

	// write each event on its own line in the file
	_, err = w.File.Write(append(byts, []byte("\n")...))
	if err != nil {
		return err
	}

	return nil
}
