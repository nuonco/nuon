package transport

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/require"

	ctlconfig "github.com/nuonco/nuon/services/ctl-api/internal"
)

type uploadMock struct {
	input   *s3.PutObjectInput
	payload []byte
	out     *manager.UploadOutput
}

func (m *uploadMock) Upload(_ context.Context, input *s3.PutObjectInput, _ ...func(*manager.Uploader)) (*manager.UploadOutput, error) {
	m.input = input
	m.payload, _ = io.ReadAll(input.Body)
	return m.out, nil
}

type getMock struct {
	input *s3.GetObjectInput
	out   *s3.GetObjectOutput
}

func (m *getMock) GetObject(_ context.Context, input *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	m.input = input
	return m.out, nil
}

type headMock struct {
	input *s3.HeadObjectInput
	out   *s3.HeadObjectOutput
	err   error
}

type deleteMock struct {
	input *s3.DeleteObjectInput
}

func (m *deleteMock) DeleteObject(_ context.Context, input *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	m.input = input
	return &s3.DeleteObjectOutput{}, nil
}

func (m *headMock) HeadObject(_ context.Context, input *s3.HeadObjectInput, _ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	m.input = input
	if m.err != nil {
		return nil, m.err
	}
	if m.out == nil {
		return nil, &types.NotFound{}
	}
	return m.out, nil
}

type presignMock struct {
	input   *s3.GetObjectInput
	expires time.Duration
}

func (m *presignMock) PresignGetObject(_ context.Context, input *s3.GetObjectInput, opts ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	m.input = input
	options := s3.PresignOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	m.expires = options.Expires
	return &v4.PresignedHTTPRequest{URL: "https://example.test/download"}, nil
}

func TestPublishNormalizesBodyAndVerifiesExactBytes(t *testing.T) {
	payload := []byte("actual fixture payload")
	digest := digestOf(payload)
	version := "version-1"
	size := int64(len(payload))
	upload := &uploadMock{out: &manager.UploadOutput{VersionID: &version}}
	get := successfulGet(payload, version, size)
	store := newS3Store("bucket", "us-west-2", "/bundles/", time.Minute, upload, get, &headMock{}, &presignMock{})
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	store.now = func() time.Time { return now }
	body := bytes.NewReader(payload)
	_, err := body.Seek(0, io.SeekEnd)
	require.NoError(t, err)

	replica, err := store.Publish(context.Background(), PublishRequest{Body: body, Size: size, SHA256: "sha256:" + digest})
	require.NoError(t, err)
	require.Equal(t, payload, upload.payload)
	require.Equal(t, "bundles/"+digest+".tar.zst", replica.StorageRef)
	require.Equal(t, version, replica.StorageVersion)
	require.Equal(t, now, replica.VerifiedAt)
	require.Equal(t, version, *get.input.VersionId)
	require.Equal(t, types.ChecksumAlgorithmSha256, upload.input.ChecksumAlgorithm)
	require.Nil(t, upload.input.ChecksumSHA256)
}

func TestDeleteRemovesExactObjectVersion(t *testing.T) {
	store := newS3Store("bucket", "us-west-2", "bundles", time.Minute, &uploadMock{}, &getMock{}, &headMock{}, &presignMock{})
	client := &deleteMock{}
	store.delete = client

	err := store.Delete(context.Background(), Replica{Provider: ProviderAWSS3, StorageRef: "bundles/archive.tar.zst", StorageVersion: "version-1"})
	require.NoError(t, err)
	require.Equal(t, "bucket", aws.ToString(client.input.Bucket))
	require.Equal(t, "bundles/archive.tar.zst", aws.ToString(client.input.Key))
	require.Equal(t, "version-1", aws.ToString(client.input.VersionId))
}

func TestPublishMultipartDoesNotSetFullObjectChecksum(t *testing.T) {
	payload := bytes.Repeat([]byte("x"), int(manager.DefaultUploadPartSize)+1)
	digest := digestOf(payload)
	version := "v1"
	size := int64(len(payload))
	upload := &uploadMock{out: &manager.UploadOutput{VersionID: &version}}
	store := newS3Store("bucket", "region", "prefix", time.Minute, upload, successfulGet(payload, version, size), &headMock{}, &presignMock{})

	_, err := store.Publish(context.Background(), PublishRequest{Body: bytes.NewReader(payload), Size: size, SHA256: digest})
	require.NoError(t, err)
	require.Nil(t, upload.input.ChecksumSHA256)
	require.Equal(t, types.ChecksumAlgorithmSha256, upload.input.ChecksumAlgorithm)
}

func TestPublishRejectsInvalidInputAndVerification(t *testing.T) {
	payload := []byte("payload")
	digest := digestOf(payload)
	version := "v1"
	tests := []struct {
		name       string
		body       []byte
		size       int64
		uploadVers *string
		getVers    string
		getBody    []byte
		getSize    int64
	}{
		{name: "declared size mismatch", body: payload, size: 1, uploadVers: &version, getVers: version, getBody: payload, getSize: int64(len(payload))},
		{name: "null upload version", body: payload, size: int64(len(payload)), uploadVers: ptr("null")},
		{name: "empty upload version", body: payload, size: int64(len(payload)), uploadVers: ptr("")},
		{name: "returned version mismatch", body: payload, size: int64(len(payload)), uploadVers: &version, getVers: "v2", getBody: payload, getSize: int64(len(payload))},
		{name: "wrong bytes", body: payload, size: int64(len(payload)), uploadVers: &version, getVers: version, getBody: []byte("xxxxxxx"), getSize: int64(len(payload))},
		{name: "truncated bytes", body: payload, size: int64(len(payload)), uploadVers: &version, getVers: version, getBody: payload[:3], getSize: int64(len(payload))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			get := successfulGet(tt.getBody, tt.getVers, tt.getSize)
			store := newS3Store("bucket", "region", "prefix", time.Minute, &uploadMock{out: &manager.UploadOutput{VersionID: tt.uploadVers}}, get, &headMock{}, &presignMock{})
			_, err := store.Publish(context.Background(), PublishRequest{Body: bytes.NewReader(tt.body), Size: tt.size, SHA256: digest})
			require.Error(t, err)
		})
	}

	store := newS3Store("bucket", "region", "prefix", time.Minute, &uploadMock{}, &getMock{}, &headMock{}, &presignMock{})
	_, err := store.Publish(context.Background(), PublishRequest{Size: -1, SHA256: digest})
	require.Error(t, err)
}

func TestNewStoreReturnsDisabledStoreWithoutBucket(t *testing.T) {
	cfg := ctlconfig.Config{BlobStorageBucket: "blob-bucket", BlobStorageRegion: "blob-region", CustomerManagedBundleGrantTTL: time.Minute}
	store, err := NewStore(S3Params{Config: &cfg})
	require.NoError(t, err)
	require.False(t, store.Configured())

	_, err = store.Publish(context.Background(), PublishRequest{})
	require.ErrorIs(t, err, ErrNotConfigured)
	_, err = store.Grant(context.Background(), Replica{}, "bundle.tar.zst", time.Time{})
	require.ErrorIs(t, err, ErrNotConfigured)
}

func TestNewS3RejectsInvalidResolvedConfiguration(t *testing.T) {
	base := ctlconfig.Config{CustomerManagedBundleStorageBucket: "bucket", CustomerManagedBundleStorageRegion: "region", CustomerManagedBundleGrantTTL: time.Minute}
	tests := []struct {
		name   string
		mutate func(*ctlconfig.Config)
	}{
		{name: "bucket", mutate: func(c *ctlconfig.Config) { c.CustomerManagedBundleStorageBucket = "" }},
		{name: "region", mutate: func(c *ctlconfig.Config) { c.CustomerManagedBundleStorageRegion = "" }},
		{name: "region not inherited from blob storage", mutate: func(c *ctlconfig.Config) {
			c.CustomerManagedBundleStorageRegion = ""
			c.BlobStorageRegion = "blob-region"
		}},
		{name: "ttl", mutate: func(c *ctlconfig.Config) { c.CustomerManagedBundleGrantTTL = 0 }},
		{name: "ttl over seven days", mutate: func(c *ctlconfig.Config) { c.CustomerManagedBundleGrantTTL = 7*24*time.Hour + time.Second }},
		{name: "endpoint scheme", mutate: func(c *ctlconfig.Config) { c.CustomerManagedBundleStorageEndpoint = "ftp://host/path" }},
		{name: "endpoint host", mutate: func(c *ctlconfig.Config) { c.CustomerManagedBundleStorageEndpoint = "https:///path" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := base
			tt.mutate(&cfg)
			_, err := NewS3(S3Params{Config: &cfg})
			require.Error(t, err)
		})
	}
}

func TestGrantPinsVersionDispositionAndEnforcesTTL(t *testing.T) {
	presign := &presignMock{}
	store := newS3Store("bucket", "region", "prefix", 15*time.Minute, &uploadMock{}, &getMock{}, &headMock{}, presign)
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	store.now = func() time.Time { return now }
	replica := Replica{Provider: ProviderAWSS3, StorageRef: "key", StorageVersion: "v7"}

	grant, err := store.Grant(context.Background(), replica, "../../bad\r\nname\".tar.zst", time.Time{})
	require.NoError(t, err)
	require.Equal(t, "v7", *presign.input.VersionId)
	require.Equal(t, `attachment; filename="bad__name_.tar.zst"`, *presign.input.ResponseContentDisposition)
	require.Equal(t, now.Add(15*time.Minute), grant.ExpiresAt)

	_, err = store.Grant(context.Background(), replica, "file", now.Add(15*time.Minute+time.Second))
	require.Error(t, err)
	_, err = store.Grant(context.Background(), Replica{Provider: ProviderAWSS3, StorageRef: "key", StorageVersion: "null"}, "file", time.Time{})
	require.Error(t, err)
}

func TestGrantWithRealPresigner(t *testing.T) {
	client := s3.NewFromConfig(aws.Config{
		Region:      "us-west-2",
		Credentials: credentials.NewStaticCredentialsProvider("access", "secret", ""),
	}, func(options *s3.Options) {
		options.BaseEndpoint = aws.String("https://objects.example.test")
		options.UsePathStyle = true
	})
	store := newS3Store("bundle-bucket", "us-west-2", "prefix", 15*time.Minute, &uploadMock{}, &getMock{}, &headMock{}, s3.NewPresignClient(client))
	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	store.now = func() time.Time { return now }

	grant, err := store.Grant(context.Background(), Replica{Provider: ProviderAWSS3, StorageRef: "prefix/key.tar.zst", StorageVersion: "exact-v1"}, "bundle.tar.zst", now.Add(10*time.Minute))
	require.NoError(t, err)
	parsed, err := url.Parse(grant.URL)
	require.NoError(t, err)
	require.Equal(t, "/bundle-bucket/prefix/key.tar.zst", parsed.Path)
	require.Equal(t, "exact-v1", parsed.Query().Get("versionId"))
	require.Equal(t, `attachment; filename="bundle.tar.zst"`, parsed.Query().Get("response-content-disposition"))
	require.NotEmpty(t, parsed.Query().Get("X-Amz-Signature"))
	require.Equal(t, "host", strings.ToLower(parsed.Query().Get("X-Amz-SignedHeaders")))
}

func TestPublishBlobUploadsAndSkipsExisting(t *testing.T) {
	payload := []byte("blob payload")
	blobDigest := digestOf(payload)
	upload := &uploadMock{out: &manager.UploadOutput{}}
	head := &headMock{}
	store := newS3Store("bucket", "us-west-2", "bundles", time.Minute, upload, &getMock{}, head, &presignMock{})

	require.NoError(t, store.PublishBlob(context.Background(), "org123", blobDigest, payload))
	require.Equal(t, "bundles/blobs/org123/sha256/"+blobDigest, *head.input.Key)
	require.Equal(t, "bundles/blobs/org123/sha256/"+blobDigest, *upload.input.Key)
	require.Equal(t, payload, upload.payload)
	require.Nil(t, upload.input.ChecksumSHA256)
	require.Equal(t, types.ChecksumAlgorithmSha256, upload.input.ChecksumAlgorithm)

	size := int64(len(payload))
	skipUpload := &uploadMock{out: &manager.UploadOutput{}}
	existing := &headMock{out: &s3.HeadObjectOutput{ContentLength: &size}}
	store = newS3Store("bucket", "us-west-2", "bundles", time.Minute, skipUpload, &getMock{}, existing, &presignMock{})
	require.NoError(t, store.PublishBlob(context.Background(), "org123", blobDigest, payload))
	require.Nil(t, skipUpload.input)
}

func TestPublishBlobRejectsInvalidInputs(t *testing.T) {
	store := newS3Store("bucket", "us-west-2", "bundles", time.Minute, &uploadMock{}, &getMock{}, &headMock{}, &presignMock{})
	require.Error(t, store.PublishBlob(context.Background(), "", digestOf([]byte("x")), []byte("x")))
	require.Error(t, store.PublishBlob(context.Background(), "org/../123", digestOf([]byte("x")), []byte("x")))
	require.Error(t, store.PublishBlob(context.Background(), "org123", "not-a-digest", []byte("x")))
}

func TestGrantBlobPresignsExistingBlobOnly(t *testing.T) {
	payload := []byte("blob payload")
	blobDigest := digestOf(payload)
	size := int64(len(payload))
	presign := &presignMock{}
	head := &headMock{out: &s3.HeadObjectOutput{ContentLength: &size}}
	store := newS3Store("bucket", "us-west-2", "bundles", time.Minute, &uploadMock{}, &getMock{}, head, presign)

	grant, err := store.GrantBlob(context.Background(), "org123", blobDigest)
	require.NoError(t, err)
	require.Equal(t, "https://example.test/download", grant.URL)
	require.Equal(t, size, grant.Size)
	require.Equal(t, time.Minute, presign.expires)
	require.Equal(t, "bundles/blobs/org123/sha256/"+blobDigest, *presign.input.Key)

	missing := newS3Store("bucket", "us-west-2", "bundles", time.Minute, &uploadMock{}, &getMock{}, &headMock{}, &presignMock{})
	_, err = missing.GrantBlob(context.Background(), "org123", blobDigest)
	require.Error(t, err)
}

func successfulGet(payload []byte, version string, size int64) *getMock {
	return &getMock{out: &s3.GetObjectOutput{VersionId: &version, ContentLength: &size, Body: io.NopCloser(bytes.NewReader(payload))}}
}

func digestOf(payload []byte) string {
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
}

func ptr(value string) *string { return &value }
