package s3

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-playground/validator/v10"
	assumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"
)

type s3Fetcher struct {
	BucketName      string `validate:"required"`
	Key             string `validate:"required"`
	RoleARN         string `validate:"required"`
	RoleSessionName string `validate:"required"`

	// internal state
	v       *validator.Validate
	fetcher fetcher
}

type s3Option func(*s3Fetcher) error

// New instantiates a new module fetcher that fetches from S3
func New(v *validator.Validate, opts ...s3Option) (*s3Fetcher, error) {
	s := &s3Fetcher{}

	if v == nil {
		return nil, fmt.Errorf("error instantiating s3 fetcher: validator is nil")
	}
	s.v = v
	s.fetcher = s

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	if err := s.v.Struct(s); err != nil {
		return nil, err
	}
	return s, nil
}

// WithBucketName sets the bucket to download from
func WithBucketName(bucketName string) s3Option {
	return func(s *s3Fetcher) error {
		s.BucketName = bucketName
		return nil
	}
}

// WithBucketKey sets the object to download
func WithBucketKey(bucketKey string) s3Option {
	return func(s *s3Fetcher) error {
		s.Key = bucketKey
		return nil
	}
}

func WithRoleARN(arn string) s3Option {
	return func(s *s3Fetcher) error {
		s.RoleARN = arn
		return nil
	}
}

func WithRoleSessionName(name string) s3Option {
	return func(s *s3Fetcher) error {
		s.RoleSessionName = name
		return nil
	}
}

// NOTE(jdt): solely for testing
type fetcher interface {
	fetch(context.Context, s3ObjectGetter) (io.ReadCloser, error)
}

// Fetch pulls the requested object from S3
func (s *s3Fetcher) Fetch(ctx context.Context) (io.ReadCloser, error) {
	assumer, err := assumerole.New(
		s.v,
		assumerole.WithRoleARN(s.RoleARN),
		assumerole.WithRoleSessionName(s.RoleSessionName),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create role assumer: %w", err)
	}

	cfg, err := assumer.LoadConfigWithAssumedRole(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to assume role: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return s.fetcher.fetch(ctx, client)
}

type s3ObjectGetter interface {
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

// fetch downloads the object from the provided S3 api
func (s *s3Fetcher) fetch(ctx context.Context, api s3ObjectGetter) (io.ReadCloser, error) {
	resp, err := api.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &s.BucketName,
		Key:    &s.Key,
	})
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
