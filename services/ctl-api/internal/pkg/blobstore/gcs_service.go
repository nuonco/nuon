package blobstore

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"cloud.google.com/go/storage"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

// gcsService is the GCS-backed implementation of Service, used when
// BLOB_STORAGE_PROVIDER=gcs. Auth is via Application Default Credentials —
// on a GCP-hosted control plane this resolves through GKE Workload Identity,
// which already carries roles/storage.admin on the blob bucket.
type gcsService struct {
	cfg    *internal.Config
	bucket *storage.BucketHandle
	mw     metrics.Writer

	// dlInFlight mirrors the S3 service's in-flight download gauge.
	dlInFlight int64
}

func newGCSService(ctx context.Context, cfg *internal.Config, mw metrics.Writer) (Service, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create gcs client: %w", err)
	}

	return &gcsService{
		cfg:    cfg,
		bucket: client.Bucket(cfg.BlobStorageBucket),
		mw:     mw,
	}, nil
}

func (s *gcsService) Upload(ctx context.Context, key string, data []byte) error {
	if _, err := s.writeObject(ctx, key, bytes.NewReader(data)); err != nil {
		return fmt.Errorf("failed to upload blob: %w", err)
	}
	return nil
}

func (s *gcsService) Delete(ctx context.Context, key string) error {
	return s.bucket.Object(key).Delete(ctx)
}

func (s *gcsService) Download(ctx context.Context, key string) ([]byte, error) {
	r, err := s.bucket.Object(key).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer r.Close()

	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read object: %w", err)
	}
	return data, nil
}

func (s *gcsService) UploadStream(ctx context.Context, key string, reader io.Reader) (string, error) {
	checksum, err := s.writeObject(ctx, key, reader)
	if err != nil {
		return "", fmt.Errorf("failed to upload stream: %w", err)
	}
	return checksum, nil
}

// writeObject streams reader into the object at key, returning a sha256:<hex>
// checksum of the content — same format as the S3 uploader.
func (s *gcsService) writeObject(ctx context.Context, key string, reader io.Reader) (string, error) {
	// Cancelling the writer's context aborts the upload on a copy error,
	// replacing the deprecated Writer.CloseWithError.
	wctx, cancel := context.WithCancel(ctx)
	defer cancel()
	w := s.bucket.Object(key).NewWriter(wctx)

	hash := sha256.New()
	if _, err := io.Copy(w, io.TeeReader(reader, hash)); err != nil {
		cancel()
		return "", fmt.Errorf("failed to write object: %w", err)
	}
	if err := w.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	return fmt.Sprintf("sha256:%x", hash.Sum(nil)), nil
}

func (s *gcsService) DownloadStream(ctx context.Context, key string) (io.ReadCloser, error) {
	inFlight := atomic.AddInt64(&s.dlInFlight, 1)
	s.mw.Gauge("blobstore.gcs.download.in_flight", float64(inFlight), nil)

	start := time.Now()
	r, err := s.bucket.Object(key).NewReader(ctx)
	if err != nil {
		s.mw.Gauge("blobstore.gcs.download.in_flight", float64(atomic.AddInt64(&s.dlInFlight, -1)), nil)
		s.mw.Timing("blobstore.gcs.get_object.latency", time.Since(start), []string{"status:error"})
		s.mw.Incr("blobstore.gcs.get_object", []string{"status:error"})
		return nil, fmt.Errorf("failed to get object: %w", err)
	}

	s.mw.Timing("blobstore.gcs.get_object.latency", time.Since(start), []string{"status:success"})
	s.mw.Incr("blobstore.gcs.get_object", []string{"status:success"})

	return &meteredBody{
		rc:    r,
		start: start,
		onClose: func(nbytes int64, dur time.Duration) {
			s.mw.Gauge("blobstore.gcs.download.in_flight", float64(atomic.AddInt64(&s.dlInFlight, -1)), nil)
			s.mw.Timing("blobstore.gcs.download.latency", dur, nil)
			s.mw.Distribution("blobstore.gcs.download.bytes", float64(nbytes), nil)
		},
	}, nil
}

func (s *gcsService) GetMetadata(ctx context.Context, key string) (int64, string, error) {
	start := time.Now()
	attrs, err := s.bucket.Object(key).Attrs(ctx)
	if err != nil {
		s.mw.Timing("blobstore.gcs.head_object.latency", time.Since(start), []string{"status:error"})
		s.mw.Incr("blobstore.gcs.head_object", []string{"status:error"})
		return 0, "", fmt.Errorf("failed to get metadata: %w", err)
	}
	s.mw.Timing("blobstore.gcs.head_object.latency", time.Since(start), []string{"status:success"})
	s.mw.Incr("blobstore.gcs.head_object", []string{"status:success"})

	contentType := attrs.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return attrs.Size, contentType, nil
}
