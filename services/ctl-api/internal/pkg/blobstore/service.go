package blobstore

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/aws/s3downloader"
	"github.com/nuonco/nuon/pkg/aws/s3uploader"
	"github.com/nuonco/nuon/pkg/metrics"
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
	s3Client   *s3.Client
	mw         metrics.Writer

	// dlInFlight tracks the number of blob downloads currently open (from
	// GetObject until the caller closes the body). Emitted as a gauge so a
	// cron burst that saturates the S3 fetch path is observable.
	dlInFlight int64
}

// NewService creates a new blob storage service
func NewService(cfg *internal.Config, mw metrics.Writer) (Service, error) {
	if cfg.BlobStorageProvider == "gcs" {
		return newGCSService(context.Background(), cfg, mw)
	}

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

	// Load AWS config for direct S3 operations. A shared HTTP client with a
	// pooled transport lets every blob download reuse TCP/TLS connections
	// instead of opening a new one per request; a fresh s3.NewFromConfig per
	// call would otherwise get its own connection pool and churn connections
	// under load.
	httpClient := awshttp.NewBuildableClient().WithTransportOptions(func(t *http.Transport) {
		t.MaxIdleConns = 200
		t.MaxIdleConnsPerHost = 200
		t.IdleConnTimeout = 90 * time.Second
	})
	awsConfig, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(cfg.BlobStorageRegion),
		awsconfig.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load aws config: %w", err)
	}

	return &service{
		cfg:        cfg,
		uploader:   uploader,
		downloader: downloader,
		s3Client:   s3.NewFromConfig(awsConfig),
		mw:         mw,
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
	inFlight := atomic.AddInt64(&s.dlInFlight, 1)
	s.mw.Gauge("blobstore.s3.download.in_flight", float64(inFlight), nil)

	start := time.Now()
	resp, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.cfg.BlobStorageBucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		// GetObject failed: nothing to read, so release the in-flight slot now.
		s.mw.Gauge("blobstore.s3.download.in_flight", float64(atomic.AddInt64(&s.dlInFlight, -1)), nil)
		s.mw.Timing("blobstore.s3.get_object.latency", time.Since(start), []string{"status:error"})
		s.mw.Incr("blobstore.s3.get_object", []string{"status:error"})
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	// GetObject latency = connection setup + time-to-first-byte, isolated from
	// the body read that the caller drives below.
	s.mw.Timing("blobstore.s3.get_object.latency", time.Since(start), []string{"status:success"})
	s.mw.Incr("blobstore.s3.get_object", []string{"status:success"})

	// The body read (and thus the bulk of the download time) happens in the
	// caller. Wrap it so the total download duration, byte count, and in-flight
	// gauge are recorded when the caller closes the stream.
	return &meteredBody{
		rc:    resp.Body,
		start: start,
		onClose: func(nbytes int64, dur time.Duration) {
			s.mw.Gauge("blobstore.s3.download.in_flight", float64(atomic.AddInt64(&s.dlInFlight, -1)), nil)
			s.mw.Timing("blobstore.s3.download.latency", dur, nil)
			s.mw.Distribution("blobstore.s3.download.bytes", float64(nbytes), nil)
		},
	}, nil
}

func (s *service) GetMetadata(ctx context.Context, s3Key string) (int64, string, error) {
	start := time.Now()
	resp, err := s.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.cfg.BlobStorageBucket),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		s.mw.Timing("blobstore.s3.head_object.latency", time.Since(start), []string{"status:error"})
		s.mw.Incr("blobstore.s3.head_object", []string{"status:error"})
		return 0, "", fmt.Errorf("failed to get metadata: %w", err)
	}
	s.mw.Timing("blobstore.s3.head_object.latency", time.Since(start), []string{"status:success"})
	s.mw.Incr("blobstore.s3.head_object", []string{"status:success"})

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

// meteredBody wraps an S3 GetObject body to record the total download duration
// and byte count when the caller closes the stream. onClose runs exactly once.
type meteredBody struct {
	rc      io.ReadCloser
	start   time.Time
	nbytes  int64
	once    sync.Once
	onClose func(nbytes int64, dur time.Duration)
}

func (m *meteredBody) Read(p []byte) (int, error) {
	n, err := m.rc.Read(p)
	m.nbytes += int64(n)
	return n, err
}

func (m *meteredBody) Close() error {
	m.once.Do(func() {
		m.onClose(m.nbytes, time.Since(m.start))
	})
	return m.rc.Close()
}
