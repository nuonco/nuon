package s3uploader

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-playground/validator/v10"
	assumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"
	"github.com/powertoolsdev/mono/pkg/aws/credentials"
)

// uploader is the interface for uploading data into output runs directory
type Uploader interface {
	// uploadFile writes the data in the file into the output s3 blob
	UploadFile(context.Context, string, string, string) error

	// uploadBlob writes the data in the byte slice into the output s3 blob
	UploadBlob(context.Context, []byte, string) error
}

func NewS3Uploader(v *validator.Validate, opts ...uploaderOptions) (*s3Uploader, error) {
	obj := &s3Uploader{
		v: v,
	}

	for _, opt := range opts {
		opt(obj)
	}
	if err := obj.v.Struct(obj); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	return obj, nil
}

type uploaderOptions func(*s3Uploader)

// WithCredentials sets the credentials
func WithCredentials(creds *credentials.Config) uploaderOptions {
	return func(obj *s3Uploader) {
		obj.creds = creds
	}
}

// WithAssumeRoleARN sets the ARN of the role to assume
func WithAssumeRoleARN(s string) uploaderOptions {
	return func(obj *s3Uploader) {
		obj.assumeRoleARN = s
	}
}

// WithAssumeSessionName sets the session name of the assume
func WithAssumeSessionName(s string) uploaderOptions {
	return func(obj *s3Uploader) {
		obj.assumeRoleSessionName = s
	}
}

// WithPrefix sets the session name of the assume
func WithPrefix(s string) uploaderOptions {
	return func(obj *s3Uploader) {
		obj.prefix = s
	}
}

// WithBucketName sets the bucket name
func WithBucketName(s string) uploaderOptions {
	return func(obj *s3Uploader) {
		obj.Bucket = s
	}
}

type s3Uploader struct {
	v *validator.Validate

	prefix string
	Bucket string `validate:"required"`

	// assumeRoleARN is an optional role which will be assumed if passed in
	assumeRoleARN         string
	assumeRoleSessionName string
	creds                 *credentials.Config
}

func (s *s3Uploader) loadAWSConfig(ctx context.Context) (aws.Config, error) {
	if s.creds != nil {
		cfg, err := credentials.Fetch(ctx, s.creds)
		if err != nil {
			return aws.Config{}, fmt.Errorf("unable to fetch credentials using config: %w", err)
		}
		return cfg, nil
	}

	if s.assumeRoleARN == "" {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return aws.Config{}, fmt.Errorf("unable to load default config: %w", err)
		}
		return cfg, nil
	}

	v := validator.New()
	assumer, err := assumerole.New(v, assumerole.WithRoleARN(s.assumeRoleARN), assumerole.WithRoleSessionName(s.assumeRoleSessionName))
	if err != nil {
		return aws.Config{}, fmt.Errorf("unable to create role assumer: %w", err)
	}
	cfg, err := assumer.LoadConfigWithAssumedRole(ctx)
	if err != nil {
		return aws.Config{}, fmt.Errorf("unable to assume role: %w", err)
	}

	return cfg, nil
}

func (s *s3Uploader) UploadFile(ctx context.Context, tmpDir, inputName, outputName string) error {
	cfg, err := s.loadAWSConfig(ctx)
	if err != nil {
		return fmt.Errorf("unable to load aws config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)

	inputFp := filepath.Join(tmpDir, inputName)
	f, err := os.Open(inputFp)
	if err != nil {
		return err
	}
	defer f.Close()

	return s.upload(ctx, uploader, f, outputName)
}

func (s *s3Uploader) UploadBlob(ctx context.Context, byts []byte, outputName string) error {
	cfg, err := s.loadAWSConfig(ctx)
	if err != nil {
		return fmt.Errorf("unable to load aws config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)
	f := bytes.NewReader(byts)

	return s.upload(ctx, uploader, f, outputName)
}

type s3UploaderClient interface {
	Upload(context.Context, *s3.PutObjectInput, ...func(*manager.Uploader)) (*manager.UploadOutput, error)
}

func (s *s3Uploader) upload(ctx context.Context, client s3UploaderClient, f io.Reader, name string) error {
	key := filepath.Join(s.prefix, name)
	bucket := s.Bucket
	_, err := client.Upload(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   f,
	})
	return err
}
