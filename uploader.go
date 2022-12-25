package uploader

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-playground/validator/v10"
	assumerole "github.com/powertoolsdev/go-aws-assume-role"
)

// TODO(jdt): add interfaces and test

// uploader is the interface for uploading data into output runs directory
type Uploader interface {
	SetUploadPrefix(string)

	// uploadFile writes the data in the file into the output s3 blob
	UploadFile(context.Context, string, string, string) error

	// uploadBlob writes the data in the byte slice into the output s3 blob
	UploadBlob(context.Context, []byte, string) error
}

func NewS3Uploader(bucket, prefix string, opts ...uploaderOptions) *s3Uploader {
	obj := &s3Uploader{
		installBucket: bucket,
		prefix:        prefix,
	}

	for _, opt := range opts {
		opt(obj)
	}

	return obj
}

type uploaderOptions func(*s3Uploader)

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

type s3Uploader struct {
	prefix        string
	installBucket string

	// assumeRoleARN is an optional role which will be assumed if passed in
	assumeRoleARN         string
	assumeRoleSessionName string
}

func (s *s3Uploader) SetUploadPrefix(prefix string) {
	ts := time.Now()
	s.prefix = filepath.Join(prefix, fmt.Sprintf("runs/ts=%d", ts.Unix()))
}

func (s *s3Uploader) loadAWSConfig(ctx context.Context) (aws.Config, error) {
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
	if s.prefix == "" {
		return fmt.Errorf("unable to upload, missing prefix")
	}

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
	if s.prefix == "" {
		return fmt.Errorf("unable to upload, missing prefix")
	}

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
	bucket := s.installBucket
	_, err := client.Upload(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   f,
	})
	return err
}
