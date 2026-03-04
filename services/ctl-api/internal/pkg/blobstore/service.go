package blobstore

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/aws/s3downloader"
	"github.com/nuonco/nuon/pkg/aws/s3uploader"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

// Service provides blob storage operations with S3
type Service interface {
	// Upload stores blob data in S3 (byte-based, for small payloads)
	Upload(ctx context.Context, s3Key string, data []byte) error

	// Download retrieves blob data from S3 (byte-based, for small payloads)
	Download(ctx context.Context, s3Key string) ([]byte, error)

	// UploadStream stores blob data in S3 (streaming, for large payloads)
	// Returns SHA256 checksum
	UploadStream(ctx context.Context, s3Key string, reader io.Reader) (checksum string, err error)

	// DownloadStream retrieves blob data from S3 (streaming, for large payloads)
	// Returns io.ReadCloser that must be closed by caller
	DownloadStream(ctx context.Context, s3Key string) (io.ReadCloser, error)

	// GetMetadata retrieves blob metadata without downloading content
	GetMetadata(ctx context.Context, s3Key string) (size int64, contentType string, err error)
}

type service struct {
	cfg        *internal.Config
	uploader   s3uploader.Uploader
	downloader s3downloader.Downloader
	awsConfig  aws.Config
}

// NewService creates a new blob storage service
func NewService(cfg *internal.Config) (Service, error) {
	v := validator.New()

	// Create uploader
	uploader, err := s3uploader.NewS3Uploader(
		v,
		s3uploader.WithBucketName(cfg.BlobStorageBucket),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create s3 uploader: %w", err)
	}

	// Create downloader
	downloader, err := s3downloader.New(
		cfg.BlobStorageBucket,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create s3 downloader: %w", err)
	}

	// Load AWS config for direct S3 operations
	awsConfig, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.BlobStorageRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	return &service{
		cfg:        cfg,
		uploader:   uploader,
		downloader: downloader,
		awsConfig:  awsConfig,
	}, nil
}

func (s *service) Upload(ctx context.Context, s3Key string, data []byte) error {
	return s.uploader.UploadBlob(ctx, data, s3Key)
}

func (s *service) Download(ctx context.Context, s3Key string) ([]byte, error) {
	return s.downloader.GetBlob(ctx, s3Key)
}

func (s *service) UploadStream(ctx context.Context, s3Key string, reader io.Reader) (string, error) {
	// UploadStream returns SHA256 checksum
	checksum, err := s.uploader.UploadStream(ctx, reader, s3Key)
	if err != nil {
		return "", fmt.Errorf("failed to upload stream: %w", err)
	}
	return checksum, nil
}

func (s *service) DownloadStream(ctx context.Context, s3Key string) (io.ReadCloser, error) {
	client := s3.NewFromConfig(s.awsConfig)

	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.BlobStorageBucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	return resp.Body, nil
}

func (s *service) GetMetadata(ctx context.Context, s3Key string) (int64, string, error) {
	client := s3.NewFromConfig(s.awsConfig)

	resp, err := client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.cfg.BlobStorageBucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return 0, "", fmt.Errorf("failed to get metadata: %w", err)
	}

	contentType := "application/octet-stream"
	if resp.ContentType != nil {
		contentType = *resp.ContentType
	}

	size := int64(0)
	if resp.ContentLength != nil {
		size = *resp.ContentLength
	}

	return size, contentType, nil
}
