package s3uploader

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	assumerole "github.com/nuonco/nuon/pkg/aws/assume-role"
	"github.com/nuonco/nuon/pkg/aws/credentials"
)

type Uploader interface {
	UploadFile(context.Context, string, string) (string, error)

	UploadBlob(context.Context, []byte, string) error

	UploadStream(context.Context, io.Reader, string) (string, error)
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

func WithCredentials(creds *credentials.Config) uploaderOptions {
	return func(obj *s3Uploader) {
		obj.creds = creds
	}
}

func WithAssumeRoleARN(s string) uploaderOptions {
	return func(obj *s3Uploader) {
		obj.assumeRoleARN = s
	}
}

func WithAssumeSessionName(s string) uploaderOptions {
	return func(obj *s3Uploader) {
		obj.assumeRoleSessionName = s
	}
}

func WithPrefix(s string) uploaderOptions {
	return func(obj *s3Uploader) {
		obj.prefix = s
	}
}

func WithBucketName(s string) uploaderOptions {
	return func(obj *s3Uploader) {
		obj.Bucket = s
	}
}

func WithUploader(uploader s3UploaderClient) uploaderOptions {
	return func(obj *s3Uploader) {
		obj.uploader = uploader
	}
}

type s3Uploader struct {
	v *validator.Validate

	prefix string
	Bucket string `validate:"required"`

	assumeRoleARN         string
	assumeRoleSessionName string
	creds                 *credentials.Config
	uploader              s3UploaderClient
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

func (s *s3Uploader) UploadFile(ctx context.Context, srcFp, outputName string) (string, error) {
	uploader, err := s.getUploader(ctx)
	if err != nil {
		return "", err
	}

	f, err := os.Open(srcFp)
	if err != nil {
		return "", err
	}
	defer f.Close()

	hash := sha256.New()
	teeReader := io.TeeReader(f, hash)

	if err := s.upload(ctx, uploader, teeReader, outputName); err != nil {
		return "", err
	}

	checksum := fmt.Sprintf("sha256:%x", hash.Sum(nil))
	return checksum, nil
}

func (s *s3Uploader) UploadBlob(ctx context.Context, byts []byte, outputName string) error {
	uploader, err := s.getUploader(ctx)
	if err != nil {
		return err
	}
	f := bytes.NewReader(byts)

	return s.upload(ctx, uploader, f, outputName)
}

func (s *s3Uploader) UploadStream(ctx context.Context, reader io.Reader, outputName string) (string, error) {
	uploader, err := s.getUploader(ctx)
	if err != nil {
		return "", err
	}

	hash := sha256.New()
	teeReader := io.TeeReader(reader, hash)

	if err := s.upload(ctx, uploader, teeReader, outputName); err != nil {
		return "", err
	}

	checksum := fmt.Sprintf("sha256:%x", hash.Sum(nil))
	return checksum, nil
}

type s3UploaderClient interface {
	Upload(context.Context, *s3.PutObjectInput, ...func(*manager.Uploader)) (*manager.UploadOutput, error)
}

func (s *s3Uploader) getUploader(ctx context.Context) (s3UploaderClient, error) {
	if s.uploader != nil {
		return s.uploader, nil
	}

	cfg, err := s.loadAWSConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load aws config: %w", err)
	}

	return manager.NewUploader(s3.NewFromConfig(cfg)), nil
}

func (s *s3Uploader) upload(ctx context.Context, client s3UploaderClient, f io.Reader, name string) error {
	key := filepath.Join(s.prefix, name)
	bucket := s.Bucket
	ctx, span := otel.Tracer("github.com/nuonco/nuon/pkg/aws/s3uploader").Start(
		ctx,
		"s3.upload",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	_, err := client.Upload(ctx, &s3.PutObjectInput{
		Bucket:            &bucket,
		Key:               &key,
		Body:              f,
		ChecksumAlgorithm: types.ChecksumAlgorithmCrc32,
	})
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
	return err
}
