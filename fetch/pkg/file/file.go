package file

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/go-playground/validator/v10"
)

type fileFetcher struct {
	File string `validate:"required,file"`

	// internal state
	validator *validator.Validate
}

type fileOption func(*fileFetcher) error

func New(v *validator.Validate, opts ...fileOption) (*fileFetcher, error) {
	f := &fileFetcher{}

	if v == nil {
		return nil, fmt.Errorf("error instantiating file fetcher: validator is nil")
	}
	f.validator = v

	for _, opt := range opts {
		if err := opt(f); err != nil {
			return nil, err
		}
	}
	if err := f.validator.Struct(f); err != nil {
		return nil, err
	}
	return f, nil
}

func WithFile(f string) fileOption {
	return func(ff *fileFetcher) error {
		ff.File = f
		return nil
	}
}

func (f *fileFetcher) Fetch(ctx context.Context) (io.ReadCloser, error) {
	return os.Open(f.File)
}
