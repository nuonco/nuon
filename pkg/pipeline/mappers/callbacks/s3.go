package callbacks

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/aws/s3uploader"
	"github.com/powertoolsdev/mono/pkg/pipeline"
)

func NewS3Callback(v *validator.Validate, opts ...s3CallbackOption) (pipeline.CallbackFn, error) {
	cb, err := newS3Callback(v, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to create s3 callback: %w", err)
	}

	return cb.callback, nil
}

func newS3Callback(v *validator.Validate, opts ...s3CallbackOption) (*s3Callback, error) {
	t := &s3Callback{v: v}
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

type s3CallbackOption func(*s3Callback) error

type s3Callback struct {
	v *validator.Validate

	Bucket       string              `validate:"required"`
	BucketPrefix string              `validate:"required"`
	Filename     string              `validate:"required"`
	Credentials  *credentials.Config `validate:"-"`
}

func (s *s3Callback) callback(ctx context.Context,
	log hclog.Logger,
	ui terminal.UI,
	byts []byte) error {
	u, err := s3uploader.NewS3Uploader(s.v,
		s3uploader.WithCredentials(s.Credentials),
		s3uploader.WithBucketName(s.Bucket),
	)
	if err != nil {
		return fmt.Errorf("unable to get uploader: %w", err)
	}

	fp := filepath.Join(s.BucketPrefix, s.Filename)
	if err := u.UploadBlob(ctx, byts, fp); err != nil {
		return fmt.Errorf("unable to upload blob: %w", err)
	}

	return nil

}

type BucketKeySettings struct {
	Bucket       string `validate:"required"`
	BucketPrefix string `validate:"required"`
	Filename     string `validate:"required"`
}

func WithBucketKeySettings(settings BucketKeySettings) s3CallbackOption {
	return func(s *s3Callback) error {
		if err := s.v.Struct(settings); err != nil {
			return fmt.Errorf("unable to validate bucket key settings: %w", err)
		}

		s.Bucket = settings.Bucket
		s.BucketPrefix = settings.BucketPrefix
		s.Filename = settings.Filename
		return nil
	}
}

func WithCredentials(creds *credentials.Config) s3CallbackOption {
	return func(s *s3Callback) error {
		if err := creds.Validate(s.v); err != nil {
			return fmt.Errorf("unable to validate credentials: %w", err)
		}

		s.Credentials = creds
		return nil
	}
}
