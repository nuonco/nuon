package callbacks

import (
	"context"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-hclog"

	"github.com/powertoolsdev/mono/pkg/pipeline"
)

func NewLocalCallback(v *validator.Validate, opts ...localCallbackOption) (pipeline.CallbackFn, error) {
	cb, err := newLocalCallback(v, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to create s3 callback: %w", err)
	}

	return cb.callback, nil
}

func newLocalCallback(v *validator.Validate, opts ...localCallbackOption) (*localCallback, error) {
	t := &localCallback{v: v}
	for _, opt := range opts {
		if err := opt(t); err != nil {
			return nil, err
		}
	}

	if err := t.v.Struct(t); err != nil {
		return nil, err
	}

	return t, nil
}

type localCallbackOption func(*localCallback) error

type localCallback struct {
	v        *validator.Validate
	Filename string `validate:"required"`
}

func (s *localCallback) callback(ctx context.Context,
	log hclog.Logger,
	byts []byte,
) error {
	// TODO: revisit this - we should accept arbitrary file paths
	fp := s.Filename

	if err := os.WriteFile(fp, byts, 0644); err != nil {
		log.Error("failed to write file", "path", fp, "error", err)
		return fmt.Errorf("failed to write file %s: %w", fp, err)
	}

	log.Info("successfully wrote file", "path", fp, "size", len(byts))
	return nil
}

func WithFilename(filename string) localCallbackOption {
	return func(s *localCallback) error {
		s.Filename = filename
		return nil
	}
}
