package transport

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"go.uber.org/fx"

	ctlconfig "github.com/nuonco/nuon/services/ctl-api/internal"
)

const ProviderAWSS3 = "aws_s3"

type uploader interface {
	Upload(context.Context, *s3.PutObjectInput, ...func(*manager.Uploader)) (*manager.UploadOutput, error)
}

type getClient interface {
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

type presigner interface {
	PresignGetObject(context.Context, *s3.GetObjectInput, ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

type deleteClient interface {
	DeleteObject(context.Context, *s3.DeleteObjectInput, ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

type S3Params struct {
	fx.In
	Config *ctlconfig.Config
}

type S3Store struct {
	bucket     string
	region     string
	prefix     string
	defaultTTL time.Duration
	uploader   uploader
	get        getClient
	delete     deleteClient
	presigner  presigner
	now        func() time.Time
}

func NewS3(params S3Params) (*S3Store, error) {
	cfg := params.Config
	if cfg == nil {
		return nil, errors.New("app bundle storage config is required")
	}
	region := cfg.AppBundleStorageRegion
	bucket := cfg.AppBundleStorageBucket
	prefix := strings.Trim(cfg.AppBundleStoragePrefix, "/")
	if prefix == "" {
		prefix = "app_bundles"
	}
	if strings.TrimSpace(bucket) == "" || strings.TrimSpace(region) == "" {
		return nil, errors.New("app bundle storage bucket and region are required")
	}
	ttl := cfg.AppBundleGrantTTL
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	if ttl > 7*24*time.Hour {
		return nil, errors.New("portable bundle grant TTL must be positive and no greater than seven days")
	}
	if endpoint := cfg.AppBundleStorageEndpoint; endpoint != "" {
		parsed, err := url.Parse(endpoint)
		if err != nil || !parsed.IsAbs() || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return nil, errors.New("app bundle storage endpoint must be an absolute HTTP(S) URL with a host")
		}
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("load AWS configuration: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.UsePathStyle = cfg.AppBundleStorageForcePathStyle
		if cfg.AppBundleStorageEndpoint != "" {
			options.BaseEndpoint = &cfg.AppBundleStorageEndpoint
		}
	})
	store := newS3Store(bucket, region, prefix, ttl, manager.NewUploader(client), client, s3.NewPresignClient(client))
	store.delete = client
	return store, nil
}

func newS3Store(bucket, region, prefix string, ttl time.Duration, upload uploader, get getClient, presign presigner) *S3Store {
	if prefix == "" {
		prefix = "app_bundles"
	}
	if ttl == 0 {
		ttl = 15 * time.Minute
	}
	return &S3Store{bucket: bucket, region: region, prefix: strings.Trim(prefix, "/"), defaultTTL: ttl, uploader: upload, get: get, presigner: presign, now: time.Now}
}

func (s *S3Store) Configured() bool { return true }

func (s *S3Store) Delete(ctx context.Context, replica Replica) error {
	if replica.Provider != ProviderAWSS3 || replica.Bucket == "" || replica.Region == "" || replica.StorageRef == "" || replica.StorageVersion == "" || replica.StorageVersion == "null" {
		return errors.New("complete aws_s3 replica with exact version is required")
	}
	if s.delete == nil {
		return errors.New("portable bundle delete client is not configured")
	}
	_, err := s.delete.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket:    &replica.Bucket,
		Key:       &replica.StorageRef,
		VersionId: &replica.StorageVersion,
	}, func(options *s3.Options) {
		options.Region = replica.Region
	})
	if err != nil {
		return fmt.Errorf("delete portable bundle: %w", err)
	}
	return nil
}

func (s *S3Store) Publish(ctx context.Context, req PublishRequest) (Replica, error) {
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
	key := path.Join(s.prefix, digest+".tar.zst")
	out, err := s.uploader.Upload(ctx, &s3.PutObjectInput{Bucket: &s.bucket, Key: &key, Body: req.Body, ContentLength: &req.Size, ContentType: stringPtr("application/zstd"), ChecksumAlgorithm: types.ChecksumAlgorithmSha256})
	if err != nil {
		return Replica{}, fmt.Errorf("upload portable bundle: %w", err)
	}
	if out.VersionID == nil || *out.VersionID == "" || *out.VersionID == "null" {
		return Replica{}, errors.New("upload returned no object version; bucket versioning is required")
	}
	replica := Replica{Provider: ProviderAWSS3, Bucket: s.bucket, Region: s.region, StorageRef: key, StorageVersion: *out.VersionID, TransportChecksum: digest, Size: req.Size}
	verifiedAt, err := s.verify(ctx, replica)
	if err != nil {
		return Replica{}, err
	}
	replica.VerifiedAt = verifiedAt
	return replica, nil
}

func (s *S3Store) verify(ctx context.Context, replica Replica) (time.Time, error) {
	if replica.StorageVersion == "" || replica.StorageVersion == "null" {
		return time.Time{}, errors.New("storage version is required")
	}
	out, err := s.get.GetObject(ctx, &s3.GetObjectInput{Bucket: &replica.Bucket, Key: &replica.StorageRef, VersionId: &replica.StorageVersion}, func(options *s3.Options) {
		options.Region = replica.Region
	})
	if err != nil {
		return time.Time{}, fmt.Errorf("verify uploaded bundle: %w", err)
	}
	if out.Body == nil {
		return time.Time{}, errors.New("get response omitted object body")
	}
	defer out.Body.Close()
	if out.VersionId == nil || *out.VersionId != replica.StorageVersion {
		return time.Time{}, errors.New("get response did not confirm the requested object version")
	}
	if out.ContentLength == nil || *out.ContentLength != replica.Size {
		return time.Time{}, fmt.Errorf("uploaded bundle size mismatch")
	}
	_, expected, err := canonicalSHA256(replica.TransportChecksum)
	if err != nil {
		return time.Time{}, err
	}
	hash := sha256.New()
	read, err := io.Copy(hash, out.Body)
	if err != nil {
		return time.Time{}, fmt.Errorf("read uploaded bundle: %w", err)
	}
	if read != replica.Size {
		return time.Time{}, errors.New("uploaded bundle byte count mismatch")
	}
	if subtle.ConstantTimeCompare(hash.Sum(nil), expected) != 1 {
		return time.Time{}, errors.New("uploaded bundle SHA-256 mismatch")
	}
	return s.now().UTC(), nil
}

func (s *S3Store) Grant(ctx context.Context, replica Replica, filename string, expiresAt time.Time) (DownloadGrant, error) {
	if replica.Provider != ProviderAWSS3 || replica.Bucket == "" || replica.Region == "" || replica.StorageRef == "" || replica.StorageVersion == "" || replica.StorageVersion == "null" {
		return DownloadGrant{}, errors.New("complete aws_s3 replica with exact version is required")
	}
	now := s.now()
	if expiresAt.IsZero() {
		expiresAt = now.Add(s.defaultTTL)
	}
	if !expiresAt.After(now) {
		return DownloadGrant{}, errors.New("grant expiry must be in the future")
	}
	if expiresAt.Sub(now) > s.defaultTTL || expiresAt.Sub(now) > 7*24*time.Hour {
		return DownloadGrant{}, errors.New("grant expiry exceeds the configured maximum TTL")
	}
	filename = safeFilename(filename)
	disposition := fmt.Sprintf("attachment; filename=%q", filename)
	out, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{Bucket: &replica.Bucket, Key: &replica.StorageRef, VersionId: &replica.StorageVersion, ResponseContentDisposition: &disposition}, func(options *s3.PresignOptions) {
		options.Expires = expiresAt.Sub(now)
		options.ClientOptions = append(options.ClientOptions, func(options *s3.Options) {
			options.Region = replica.Region
		})
	})
	if err != nil {
		return DownloadGrant{}, fmt.Errorf("presign portable bundle: %w", err)
	}
	return DownloadGrant{URL: out.URL, ExpiresAt: expiresAt, SupportsRange: true}, nil
}

func stringPtr(value string) *string { return &value }
