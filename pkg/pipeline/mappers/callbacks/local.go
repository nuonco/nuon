package callbacks

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-hclog"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/pipeline"
)

func NewLocalCallback(v *validator.Validate, opts ...localCallbackOption) (pipeline.CallbackFn, error) {
	cb, err := newLocalCallback(v, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to create local callback: %w", err)
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
	Compress bool   // if we should ADDITIONALLY gzip
}

// TODO(fd): this should likely be a standalone mapper to avoid overloading this method
func (s *localCallback) callback(ctx context.Context,
	log hclog.Logger,
	byts []byte,
) error {
	// TODO: revisit this - we should accept arbitrary file paths
	fp := s.Filename

	err := os.WriteFile(fp, byts, 0644)
	if err != nil {
		log.Error("failed to write file", "path", fp, "error", err)
		return fmt.Errorf("failed to write file %s: %w", fp, err)
	}
	log.Info("successfully wrote file", zap.String("path", fp), zap.Int("size", len(byts)))

	if s.Compress {
		filename := s.Filename + ".gz"
		var zipBytes bytes.Buffer
		gzipWriter := gzip.NewWriter(&zipBytes)
		gzipWriter.Write(byts)
		gzipWriter.Close()

		log.Debug("writing compressed file", zap.String("path", filename), zap.Int("file.bytes", len(byts)))
		cfd, err := os.Create(filename)
		if err != nil {
			defer cfd.Close()
			return fmt.Errorf("failed to create file %s: %w", filename, err)
		}
		n, err := cfd.Write(zipBytes.Bytes())
		if err != nil {
			return fmt.Errorf("failed to write to file %s: %w", filename, err)
		}
		log.Debug("wrote compressed file", zap.String("path", filename), zap.Int("file.bytes", len(zipBytes.Bytes())), zap.Int("bytes-written", n))
		cfd.Sync()

	}

	return nil
}

func WithFilename(filename string) localCallbackOption {
	return func(s *localCallback) error {
		s.Filename = filename
		return nil
	}
}

func WithCompression(compress bool) localCallbackOption {
	return func(s *localCallback) error {
		s.Compress = compress
		return nil
	}
}
