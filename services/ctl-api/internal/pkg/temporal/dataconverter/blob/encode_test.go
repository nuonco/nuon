package blob

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"testing"
	"time"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/filecache"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type countingBlobSvc struct {
	uploads   int
	lastS3Key string
}

func (c *countingBlobSvc) Upload(ctx context.Context, s3Key string, data []byte) error { return nil }
func (c *countingBlobSvc) Download(ctx context.Context, s3Key string) ([]byte, error) {
	return nil, nil
}
func (c *countingBlobSvc) UploadStream(ctx context.Context, s3Key string, reader io.Reader) (string, error) {
	c.uploads++
	c.lastS3Key = s3Key
	_, _ = io.Copy(io.Discard, reader)
	return "checksum", nil
}
func (c *countingBlobSvc) DownloadStream(ctx context.Context, s3Key string) (io.ReadCloser, error) {
	return nil, nil
}
func (c *countingBlobSvc) GetMetadata(ctx context.Context, s3Key string) (int64, string, error) {
	return 0, "", nil
}

type noopMetrics struct{}

func (noopMetrics) Incr(string, []string)                  {}
func (noopMetrics) Decr(string, []string)                  {}
func (noopMetrics) Timing(string, time.Duration, []string) {}
func (noopMetrics) Gauge(string, float64, []string)        {}
func (noopMetrics) Count(string, int64, []string)          {}
func (noopMetrics) Distribution(string, float64, []string) {}
func (noopMetrics) Event(*statsd.Event)                    {}
func (noopMetrics) Flush()                                 {}

func newTestConverter(t *testing.T, blobSvc *countingBlobSvc) *dataConverter {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec("CREATE TABLE temporal_blobs (id text PRIMARY KEY, created_at datetime, updated_at datetime, deleted_at integer, s3_key text NOT NULL, checksum text, size integer)").Error)

	cache, err := filecache.New(filecache.Options{Dir: t.TempDir(), MaxCount: 100, MaxBytes: 10 << 20})
	require.NoError(t, err)

	return &dataConverter{
		cfg: &internal.Config{
			TemporalDataConverterLargePayloadSize: 10,
			TemporalBlobS3Prefix:                  "prefix/",
			TemporalBlobS3Timeout:                 5 * time.Second,
		},
		l:             zap.NewNop(),
		db:            db,
		blobSvc:       blobSvc,
		mw:            noopMetrics{},
		cache:         cache,
		encodeEnabled: true,
	}
}

func newPayload(data []byte) *commonpb.Payload {
	return &commonpb.Payload{
		Metadata: map[string][]byte{converter.MetadataEncoding: []byte("json/plain")},
		Data:     data,
	}
}

func TestEncodeContentAddressedKey(t *testing.T) {
	blobSvc := &countingBlobSvc{}
	d := newTestConverter(t, blobSvc)

	data := []byte("this payload is larger than the threshold")
	encoded, err := d.encodePayload(newPayload(data))
	require.NoError(t, err)

	sum := sha256.Sum256(data)
	wantID := blobIDPrefix + hex.EncodeToString(sum[:])

	assert.Equal(t, wantID, string(encoded.Data))
	assert.Equal(t, wantID, string(encoded.Metadata["nuon/blob/blob_id"]))
	assert.Equal(t, "prefix/"+wantID, string(encoded.Metadata["nuon/blob/s3_key"]))
	assert.Equal(t, []byte(encoding), encoded.Metadata[converter.MetadataEncoding])
}

func TestEncodeDedupesIdenticalContent(t *testing.T) {
	blobSvc := &countingBlobSvc{}
	d := newTestConverter(t, blobSvc)

	data := []byte("this payload is larger than the threshold")

	first, err := d.encodePayload(newPayload(data))
	require.NoError(t, err)
	second, err := d.encodePayload(newPayload(data))
	require.NoError(t, err)

	assert.Equal(t, 1, blobSvc.uploads, "identical content must upload only once")
	assert.Equal(t, string(first.Data), string(second.Data), "identical content must map to the same key")
}

func TestEncodeDistinctContentUploadsSeparately(t *testing.T) {
	blobSvc := &countingBlobSvc{}
	d := newTestConverter(t, blobSvc)

	a, err := d.encodePayload(newPayload([]byte("payload A over the threshold")))
	require.NoError(t, err)
	b, err := d.encodePayload(newPayload([]byte("payload B over the threshold")))
	require.NoError(t, err)

	assert.Equal(t, 2, blobSvc.uploads, "distinct content must upload separately")
	assert.NotEqual(t, string(a.Data), string(b.Data), "distinct content must map to different keys")
}
