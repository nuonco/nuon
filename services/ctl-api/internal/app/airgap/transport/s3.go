package transport

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
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
	presigner  presigner
	now        func() time.Time
}

// NewStore returns a disabled store when no bucket is configured: air-gap
// bundle publishing and downloads are opt-in per deployment and must never
// silently fall back to the shared blob-storage bucket.
func NewStore(params S3Params) (Store, error) {
	if params.Config == nil {
		return nil, errors.New("air-gap bundle storage config is required")
	}
	if strings.TrimSpace(params.Config.AirgapBundleStorageBucket) == "" {
		return NewDisabled(), nil
	}
	return NewS3(params)
}

func NewS3(params S3Params) (*S3Store, error) {
	cfg := params.Config
	if cfg == nil {
		return nil, errors.New("air-gap bundle storage config is required")
	}
	region := cfg.AirgapBundleStorageRegion
	bucket := cfg.AirgapBundleStorageBucket
	prefix := strings.Trim(cfg.AirgapBundleStoragePrefix, "/")
	if prefix == "" {
		prefix = "airgap-bundles"
	}
	if strings.TrimSpace(bucket) == "" || strings.TrimSpace(region) == "" {
		return nil, errors.New("air-gap bundle storage bucket and region are required")
	}
	if cfg.AirgapBundleGrantTTL <= 0 || cfg.AirgapBundleGrantTTL > 7*24*time.Hour {
		return nil, errors.New("air-gap bundle grant TTL must be positive and no greater than seven days")
	}
	if endpoint := cfg.AirgapBundleStorageEndpoint; endpoint != "" {
		parsed, err := url.Parse(endpoint)
		if err != nil || !parsed.IsAbs() || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return nil, errors.New("air-gap bundle storage endpoint must be an absolute HTTP(S) URL with a host")
		}
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("load AWS configuration: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.UsePathStyle = cfg.AirgapBundleStorageForcePathStyle
		if cfg.AirgapBundleStorageEndpoint != "" {
			options.BaseEndpoint = &cfg.AirgapBundleStorageEndpoint
		}
	})
	return newS3Store(bucket, region, prefix, cfg.AirgapBundleGrantTTL, manager.NewUploader(client), client, s3.NewPresignClient(client)), nil
}

func newS3Store(bucket, region, prefix string, ttl time.Duration, upload uploader, get getClient, presign presigner) *S3Store {
	if prefix == "" {
		prefix = "airgap-bundles"
	}
	if ttl == 0 {
		ttl = 15 * time.Minute
	}
	return &S3Store{bucket: bucket, region: region, prefix: strings.Trim(prefix, "/"), defaultTTL: ttl, uploader: upload, get: get, presigner: presign, now: time.Now}
}

func (s *S3Store) Configured() bool { return true }

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
		return Replica{}, fmt.Errorf("upload air-gap bundle: %w", err)
	}
	if out.VersionID == nil || *out.VersionID == "" || *out.VersionID == "null" {
		return Replica{}, errors.New("upload returned no object version; bucket versioning is required")
	}
	replica := Replica{Provider: ProviderAWSS3, Region: s.region, StorageRef: key, StorageVersion: *out.VersionID, TransportChecksum: digest, Size: req.Size}
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
	out, err := s.get.GetObject(ctx, &s3.GetObjectInput{Bucket: &s.bucket, Key: &replica.StorageRef, VersionId: &replica.StorageVersion})
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
	if replica.Provider != ProviderAWSS3 || replica.StorageRef == "" || replica.StorageVersion == "" || replica.StorageVersion == "null" {
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
	out, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{Bucket: &s.bucket, Key: &replica.StorageRef, VersionId: &replica.StorageVersion, ResponseContentDisposition: &disposition}, func(options *s3.PresignOptions) { options.Expires = expiresAt.Sub(now) })
	if err != nil {
		return DownloadGrant{}, fmt.Errorf("presign air-gap bundle: %w", err)
	}
	return DownloadGrant{URL: out.URL, ExpiresAt: expiresAt, SupportsRange: true}, nil
}

func canonicalSHA256(value string) (string, []byte, error) {
	value = strings.TrimPrefix(strings.ToLower(value), "sha256:")
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != sha256.Size {
		return "", nil, errors.New("SHA-256 must be 64 hexadecimal characters")
	}
	return value, decoded, nil
}

func safeFilename(value string) string {
	value = path.Base(strings.ReplaceAll(value, "\\", "/"))
	value = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f || r == '"' {
			return '_'
		}
		return r
	}, value)
	if value == "" || value == "." {
		return "airgap-bundle.tar.zst"
	}
	return value
}

func stringPtr(value string) *string { return &value }
