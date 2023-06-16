package s3downloader

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=downloader_mock.go -source=downloader.go -package=s3downloader
type Downloader interface {
	GetBlob(context.Context, string) ([]byte, error)

	ListPrefix(context.Context, string) ([]string, error)
	ListAll(context.Context) ([]string, error)
}

// s3Downloader implements the downloader interface and exposes the abilty to get and list prefixes
type s3Downloader struct {
	v *validator.Validate `validate:"required"`

	Bucket      string              `validate:"required"`
	Credentials *credentials.Config `validate:"-"`
}

var _ Downloader = (*s3Downloader)(nil)

type downloaderOption func(*s3Downloader) error

func WithCredentials(cfg *credentials.Config) downloaderOption {
	return func(s *s3Downloader) error {
		s.Credentials = cfg
		return nil
	}
}

func New(bucket string, opts ...downloaderOption) (*s3Downloader, error) {
	dl := &s3Downloader{
		v:      validator.New(),
		Bucket: bucket,
	}
	for idx, opt := range opts {
		if err := opt(dl); err != nil {
			return nil, fmt.Errorf("error occurred during opt: %v: %w", idx, err)
		}
	}
	if err := dl.v.Struct(dl); err != nil {
		return nil, fmt.Errorf("unable to validate downloader: %w", err)
	}

	return dl, nil
}
