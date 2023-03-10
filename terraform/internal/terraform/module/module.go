package module

import (
	"context"
	"fmt"
	"io"

	"github.com/go-playground/validator/v10"
	"github.com/mholt/archiver/v4"
)

type writeFactory interface {
	GetWriter(string) (io.WriteCloser, error)
}

type fetcher interface {
	Fetch(context.Context) (io.ReadCloser, error)
}

type module struct {
	Fetcher      fetcher      `validate:"required"`
	WriteFactory writeFactory `validate:"required"`

	// internal state
	validator *validator.Validate
}

type moduleOption func(*module) error

// New instantiates a new module with the passed fetcher and writeFactory
func New(v *validator.Validate, opts ...moduleOption) (*module, error) {
	s := &module{}

	if v == nil {
		return nil, fmt.Errorf("error instantiating module: validator is nil")
	}
	s.validator = v

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	if err := s.validator.Struct(s); err != nil {
		return nil, err
	}
	return s, nil
}

// WithFetcher sets the object to download to bucketName
func WithFetcher(fetcher fetcher) moduleOption {
	return func(s *module) error {
		s.Fetcher = fetcher
		return nil
	}
}

// WithWriteFactory sets the factory used to write the fetched files
// typically provided by the workspace manager
func WithWriteFactory(w writeFactory) moduleOption {
	return func(s *module) error {
		s.WriteFactory = w
		return nil
	}
}

func (s *module) Install(ctx context.Context) error {
	iorc, err := s.Fetcher.Fetch(ctx)
	if err != nil {
		return err
	}
	defer iorc.Close()

	return s.extractModule(ctx, iorc)
}

// extractModule: accepts a gzipped, tar'd module and extracts it into the tmpdir
func (s *module) extractModule(ctx context.Context, r io.Reader) error {
	gz := archiver.Gz{}
	reader, err := gz.OpenReader(r)
	if err != nil {
		return err
	}
	defer reader.Close()

	tar := archiver.Tar{}
	if err := tar.Extract(ctx, reader, nil, func(ctx context.Context, f archiver.File) error {
		inputFile, err := f.Open()
		if err != nil {
			return err
		}
		defer inputFile.Close()

		outputFile, err := s.WriteFactory.GetWriter(f.NameInArchive)
		if err != nil {
			return err
		}
		if outputFile == nil {
			return nil
		}
		defer outputFile.Close()

		_, err = io.Copy(outputFile, inputFile)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
