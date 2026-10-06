package transport

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"go.uber.org/fx"
)

const ProviderGCS = "gcs"

type GCSStore struct {
	client      *storage.Client
	bucket      string
	region      string
	prefix      string
	signerEmail string
	defaultTTL  time.Duration
	now         func() time.Time
}

func NewGCS(params StoreParams) (*GCSStore, error) {
	cfg := params.Config
	if cfg == nil || strings.TrimSpace(cfg.AppBundleStorageBucket) == "" {
		return nil, errors.New("app bundle storage bucket is required")
	}
	if cfg.AppBundleStorageEndpoint != "" || cfg.AppBundleStorageForcePathStyle {
		return nil, errors.New("S3 endpoint and path-style settings are not supported by native GCS bundle storage")
	}
	ttl := cfg.AppBundleGrantTTL
	if ttl == 0 {
		ttl = 15 * time.Minute
	}
	if ttl < time.Second || ttl > 7*24*time.Hour {
		return nil, errors.New("portable bundle grant TTL must be at least one second and no greater than seven days")
	}
	prefix := strings.Trim(cfg.AppBundleStoragePrefix, "/")
	if prefix == "" {
		prefix = "app_bundles"
	}
	client, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("create GCS bundle storage client: %w", err)
	}
	if params.Lifecycle != nil {
		params.Lifecycle.Append(fx.Hook{OnStop: func(context.Context) error { return client.Close() }})
	}
	return &GCSStore{
		client: client, bucket: cfg.AppBundleStorageBucket, region: cfg.AppBundleStorageRegion,
		prefix: prefix, signerEmail: cfg.AppBundleStorageGCSServiceAccount, defaultTTL: ttl, now: time.Now,
	}, nil
}

func (s *GCSStore) Configured() bool { return true }

func (s *GCSStore) Publish(ctx context.Context, req PublishRequest) (Replica, error) {
	if req.Body == nil || req.Size < 0 {
		return Replica{}, errors.New("publish body and non-negative size are required")
	}
	size, err := req.Body.Seek(0, io.SeekEnd)
	if err != nil {
		return Replica{}, fmt.Errorf("determine publish body size: %w", err)
	}
	if size != req.Size {
		return Replica{}, fmt.Errorf("publish body size %d does not match declared size %d", size, req.Size)
	}
	if _, err := req.Body.Seek(0, io.SeekStart); err != nil {
		return Replica{}, fmt.Errorf("rewind publish body: %w", err)
	}
	digest, _, err := canonicalSHA256(req.SHA256)
	if err != nil {
		return Replica{}, fmt.Errorf("invalid publish request: %w", err)
	}
	bucket := s.client.Bucket(s.bucket)
	attrs, err := bucket.Attrs(ctx)
	if err != nil {
		return Replica{}, fmt.Errorf("check bundle bucket versioning: %w", err)
	}
	if !attrs.VersioningEnabled {
		return Replica{}, errors.New("bundle bucket versioning is required")
	}
	key := path.Join(s.prefix, digest+".tar.zst")
	written, err := s.writeObject(ctx, bucket.Object(key), req.Body, req.Size, digest, "application/zstd")
	if err != nil {
		return Replica{}, fmt.Errorf("upload portable bundle: %w", err)
	}
	replica := Replica{
		Provider: ProviderGCS, Bucket: s.bucket, Region: s.region, StorageRef: key,
		StorageVersion: strconv.FormatInt(written.Generation, 10), TransportChecksum: digest, Size: req.Size,
	}
	if err := s.verify(ctx, replica); err != nil {
		return Replica{}, err
	}
	replica.VerifiedAt = s.now().UTC()
	return replica, nil
}

func (s *GCSStore) writeObject(ctx context.Context, object *storage.ObjectHandle, body io.Reader, size int64, digest, contentType string) (*storage.ObjectAttrs, error) {
	_, expected, err := canonicalSHA256(digest)
	if err != nil {
		return nil, err
	}
	wctx, cancel := context.WithCancel(ctx)
	defer cancel()
	w := object.NewWriter(wctx)
	w.ContentType = contentType
	w.CacheControl = "no-store"
	w.Metadata = map[string]string{"sha256": digest}
	hash := sha256.New()
	written, err := io.Copy(w, io.TeeReader(body, hash))
	if err != nil {
		return nil, fmt.Errorf("write object: %w", err)
	}
	if written != size || subtle.ConstantTimeCompare(hash.Sum(nil), expected) != 1 {
		return nil, errors.New("object bytes do not match the declared size and SHA-256")
	}
	if err := w.Close(); err != nil {
		return nil, fmt.Errorf("complete object upload: %w", err)
	}
	attrs := w.Attrs()
	if attrs == nil || attrs.Generation <= 0 || attrs.Size != size {
		return nil, errors.New("upload returned no exact object generation or an incorrect size")
	}
	return attrs, nil
}

func gcsGeneration(replica Replica) (int64, error) {
	if replica.Provider != ProviderGCS || replica.Bucket == "" || replica.StorageRef == "" {
		return 0, errors.New("complete gcs replica with exact generation is required")
	}
	generation, err := strconv.ParseInt(replica.StorageVersion, 10, 64)
	if err != nil || generation <= 0 {
		return 0, errors.New("a positive GCS object generation is required")
	}
	return generation, nil
}

func (s *GCSStore) verify(ctx context.Context, replica Replica) error {
	generation, err := gcsGeneration(replica)
	if err != nil {
		return err
	}
	reader, err := s.client.Bucket(replica.Bucket).Object(replica.StorageRef).Generation(generation).NewReader(ctx)
	if err != nil {
		return fmt.Errorf("verify uploaded bundle: %w", err)
	}
	defer reader.Close()
	if reader.Attrs.Generation != generation || reader.Attrs.Size != replica.Size {
		return errors.New("get response did not confirm the requested object generation and size")
	}
	_, expected, err := canonicalSHA256(replica.TransportChecksum)
	if err != nil {
		return err
	}
	hash := sha256.New()
	read, err := io.Copy(hash, reader)
	if err != nil {
		return fmt.Errorf("read uploaded bundle: %w", err)
	}
	if read != replica.Size || subtle.ConstantTimeCompare(hash.Sum(nil), expected) != 1 {
		return errors.New("uploaded bundle byte count or SHA-256 mismatch")
	}
	return nil
}

func (s *GCSStore) Delete(ctx context.Context, replica Replica) error {
	generation, err := gcsGeneration(replica)
	if err != nil {
		return err
	}
	if err := s.client.Bucket(replica.Bucket).Object(replica.StorageRef).Generation(generation).Delete(ctx); err != nil {
		return fmt.Errorf("delete portable bundle: %w", err)
	}
	return nil
}

func (s *GCSStore) Grant(ctx context.Context, replica Replica, filename string, expiresAt time.Time) (DownloadGrant, error) {
	generation, err := gcsGeneration(replica)
	if err != nil {
		return DownloadGrant{}, err
	}
	if err := ctx.Err(); err != nil {
		return DownloadGrant{}, err
	}
	now := s.now()
	if expiresAt.IsZero() {
		expiresAt = now.Add(s.defaultTTL)
	}
	if expiresAt.Sub(now) < time.Second || expiresAt.Sub(now) > s.defaultTTL || expiresAt.Sub(now) > 7*24*time.Hour {
		return DownloadGrant{}, errors.New("grant expiry must be at least one second in the future and within the configured maximum TTL")
	}
	query := url.Values{
		"generation":                   {strconv.FormatInt(generation, 10)},
		"response-content-disposition": {fmt.Sprintf("attachment; filename=%q", safeFilename(filename))},
	}
	url, err := s.client.Bucket(replica.Bucket).SignedURL(replica.StorageRef, &storage.SignedURLOptions{
		Scheme: storage.SigningSchemeV4, Method: http.MethodGet, Expires: expiresAt,
		GoogleAccessID: s.signerEmail, QueryParameters: query,
	})
	if err != nil {
		return DownloadGrant{}, fmt.Errorf("sign portable bundle download: %w", err)
	}
	return DownloadGrant{URL: url, ExpiresAt: expiresAt, SupportsRange: true}, nil
}
