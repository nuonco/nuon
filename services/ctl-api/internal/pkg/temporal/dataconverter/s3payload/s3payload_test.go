package s3payload

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/sdk/converter"
	"go.uber.org/zap"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=s3payload_mock_test.go -source=s3payload_test.go -package=s3payload

// MockBlobService is the interface for mocking blobstore operations
type MockBlobService interface {
	UploadStream(ctx context.Context, s3Key string, reader io.Reader) (checksum string, err error)
	DownloadStream(ctx context.Context, s3Key string) (io.ReadCloser, error)
}

// mockMetricsWriter implements metrics.Writer for testing
type mockMetricsWriter struct{}

func (m *mockMetricsWriter) Incr(name string, tags []string)                        {}
func (m *mockMetricsWriter) Decr(name string, tags []string)                        {}
func (m *mockMetricsWriter) Timing(name string, value time.Duration, tags []string) {}
func (m *mockMetricsWriter) Gauge(name string, value float64, tags []string)        {}
func (m *mockMetricsWriter) Count(name string, value int64, tags []string)          {}
func (m *mockMetricsWriter) Event(e *statsd.Event)                                  {}
func (m *mockMetricsWriter) Flush()                                                 {}

// mockReadCloser wraps a strings.Reader to implement io.ReadCloser
type mockReadCloser struct {
	*strings.Reader
}

func (m *mockReadCloser) Close() error {
	return nil
}

func newMockReadCloser(data string) io.ReadCloser {
	return &mockReadCloser{
		Reader: strings.NewReader(data),
	}
}

// mockBlobService is a simple mock implementation for testing
type mockBlobService struct {
	uploadFn   func(ctx context.Context, s3Key string, reader io.Reader) (string, error)
	downloadFn func(ctx context.Context, s3Key string) (io.ReadCloser, error)
	storage    map[string]string // Simple in-memory storage for round-trip tests
}

func newMockBlobService() *mockBlobService {
	return &mockBlobService{
		storage: make(map[string]string),
	}
}

func (m *mockBlobService) Upload(ctx context.Context, s3Key string, data []byte) error {
	m.storage[s3Key] = string(data)
	return nil
}

func (m *mockBlobService) Download(ctx context.Context, s3Key string) ([]byte, error) {
	data, ok := m.storage[s3Key]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", s3Key)
	}
	return []byte(data), nil
}

func (m *mockBlobService) UploadStream(ctx context.Context, s3Key string, reader io.Reader) (string, error) {
	if m.uploadFn != nil {
		return m.uploadFn(ctx, s3Key, reader)
	}

	// Default: store in memory
	data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	m.storage[s3Key] = string(data)
	return "sha256:mockedchecksum", nil
}

func (m *mockBlobService) DownloadStream(ctx context.Context, s3Key string) (io.ReadCloser, error) {
	if m.downloadFn != nil {
		return m.downloadFn(ctx, s3Key)
	}

	// Default: retrieve from memory
	data, ok := m.storage[s3Key]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", s3Key)
	}
	return newMockReadCloser(data), nil
}

func (m *mockBlobService) GetMetadata(ctx context.Context, s3Key string) (int64, string, error) {
	data, ok := m.storage[s3Key]
	if !ok {
		return 0, "", fmt.Errorf("key not found: %s", s3Key)
	}
	return int64(len(data)), "application/octet-stream", nil
}

func TestDataConverter_Encode(t *testing.T) {
	logger := zap.NewNop()
	mw := &mockMetricsWriter{}

	smallPayload := []byte("small payload")
	largePayload := make([]byte, 2000) // Above threshold
	for i := range largePayload {
		largePayload[i] = byte('A' + (i % 26))
	}

	tests := map[string]struct {
		payloads    []*commonpb.Payload
		cfg         *internal.Config
		blobService *mockBlobService
		assertFn    func(*testing.T, []*commonpb.Payload, error)
	}{
		"small payload - skip encoding": {
			cfg: &internal.Config{
				TemporalDataConverterLargePayloadSize: 1000,
			},
			payloads: []*commonpb.Payload{
				{
					Metadata: map[string][]byte{
						converter.MetadataEncoding: []byte("json"),
					},
					Data: smallPayload,
				},
			},
			blobService: newMockBlobService(),
			assertFn: func(t *testing.T, result []*commonpb.Payload, err error) {
				require.NoError(t, err)
				require.Len(t, result, 1)

				// Should return original payload unchanged
				assert.Equal(t, smallPayload, result[0].Data)
				assert.Equal(t, []byte("json"), result[0].Metadata[converter.MetadataEncoding])
			},
		},
		"large payload - encode to S3": {
			cfg: &internal.Config{
				TemporalDataConverterLargePayloadSize: 1000,
			},
			payloads: []*commonpb.Payload{
				{
					Metadata: map[string][]byte{
						converter.MetadataEncoding: []byte("json"),
					},
					Data: largePayload,
				},
			},
			blobService: newMockBlobService(),
			assertFn: func(t *testing.T, result []*commonpb.Payload, err error) {
				require.NoError(t, err)
				require.Len(t, result, 1)

				// Should be S3 encoded
				assert.Equal(t, []byte("nuon/s3payload"), result[0].Metadata[converter.MetadataEncoding])
				assert.Equal(t, []byte("true"), result[0].Metadata["nuon/s3payload/enabled"])

				// Should have S3 key in data
				s3Key := string(result[0].Data)
				assert.NotEmpty(t, s3Key)
				assert.Contains(t, s3Key, "temporal/") // Default org

				// Should preserve original encoding
				assert.Equal(t, []byte("json"), result[0].Metadata["nuon/s3payload/original-encoding"])
			},
		},
		"already encoded - skip": {
			cfg: &internal.Config{
				TemporalDataConverterLargePayloadSize: 1000,
			},
			payloads: []*commonpb.Payload{
				{
					Metadata: map[string][]byte{
						converter.MetadataEncoding: []byte("nuon/s3payload"),
					},
					Data: []byte("temporal/blbabc123"),
				},
			},
			blobService: newMockBlobService(),
			assertFn: func(t *testing.T, result []*commonpb.Payload, err error) {
				require.NoError(t, err)
				require.Len(t, result, 1)

				// Should return unchanged
				assert.Equal(t, []byte("temporal/blbabc123"), result[0].Data)
			},
		},
		"S3 upload error - graceful degradation": {
			cfg: &internal.Config{
				TemporalDataConverterLargePayloadSize: 1000,
			},
			payloads: []*commonpb.Payload{
				{
					Metadata: map[string][]byte{
						converter.MetadataEncoding: []byte("json"),
					},
					Data: largePayload,
				},
			},
			blobService: &mockBlobService{
				uploadFn: func(ctx context.Context, s3Key string, reader io.Reader) (string, error) {
					return "", fmt.Errorf("S3 upload failed")
				},
			},
			assertFn: func(t *testing.T, result []*commonpb.Payload, err error) {
				require.NoError(t, err)
				require.Len(t, result, 1)

				// Should return original payload on error
				assert.Equal(t, largePayload, result[0].Data)
				assert.Equal(t, []byte("json"), result[0].Metadata[converter.MetadataEncoding])
			},
		},
		"multiple payloads - mixed sizes": {
			cfg: &internal.Config{
				TemporalDataConverterLargePayloadSize: 1000,
			},
			payloads: []*commonpb.Payload{
				{
					Metadata: map[string][]byte{converter.MetadataEncoding: []byte("json")},
					Data:     smallPayload,
				},
				{
					Metadata: map[string][]byte{converter.MetadataEncoding: []byte("json")},
					Data:     largePayload,
				},
			},
			blobService: newMockBlobService(),
			assertFn: func(t *testing.T, result []*commonpb.Payload, err error) {
				require.NoError(t, err)
				require.Len(t, result, 2)

				// First should be unchanged (small)
				assert.Equal(t, smallPayload, result[0].Data)
				assert.Equal(t, []byte("json"), result[0].Metadata[converter.MetadataEncoding])

				// Second should be S3 encoded (large)
				assert.Equal(t, []byte("nuon/s3payload"), result[1].Metadata[converter.MetadataEncoding])
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			dc := &dataConverter{
				cfg:     test.cfg,
				l:       logger,
				blobSvc: test.blobService,
				mw:      mw,
			}

			result, err := dc.Encode(test.payloads)
			test.assertFn(t, result, err)
		})
	}
}

func TestDataConverter_Decode(t *testing.T) {
	logger := zap.NewNop()
	mw := &mockMetricsWriter{}

	originalData := []byte("this is the original large payload data")

	tests := map[string]struct {
		payloads    []*commonpb.Payload
		cfg         *internal.Config
		blobService *mockBlobService
		assertFn    func(*testing.T, []*commonpb.Payload, error)
	}{
		"not S3 encoded - return as-is": {
			cfg: &internal.Config{},
			payloads: []*commonpb.Payload{
				{
					Metadata: map[string][]byte{
						converter.MetadataEncoding: []byte("json"),
					},
					Data: []byte("regular data"),
				},
			},
			blobService: newMockBlobService(),
			assertFn: func(t *testing.T, result []*commonpb.Payload, err error) {
				require.NoError(t, err)
				require.Len(t, result, 1)

				assert.Equal(t, []byte("regular data"), result[0].Data)
			},
		},
		"S3 encoded - decode successfully": {
			cfg: &internal.Config{},
			payloads: []*commonpb.Payload{
				{
					Metadata: map[string][]byte{
						converter.MetadataEncoding:         []byte("nuon/s3payload"),
						"nuon/s3payload/enabled":           []byte("true"),
						"nuon/s3payload/s3_key":            []byte("temporal/blbabc123"),
						"nuon/s3payload/original-encoding": []byte("json"),
					},
					Data: []byte("temporal/blbabc123"),
				},
			},
			blobService: &mockBlobService{
				storage: map[string]string{
					"temporal/blbabc123": string(originalData),
				},
			},
			assertFn: func(t *testing.T, result []*commonpb.Payload, err error) {
				require.NoError(t, err)
				require.Len(t, result, 1)

				// Should have original data restored
				assert.Equal(t, originalData, result[0].Data)

				// Should restore original encoding
				assert.Equal(t, []byte("json"), result[0].Metadata[converter.MetadataEncoding])
			},
		},
		"S3 download error": {
			cfg: &internal.Config{},
			payloads: []*commonpb.Payload{
				{
					Metadata: map[string][]byte{
						converter.MetadataEncoding: []byte("nuon/s3payload"),
						"nuon/s3payload/enabled":   []byte("true"),
					},
					Data: []byte("temporal/nonexistent"),
				},
			},
			blobService: newMockBlobService(), // Empty storage
			assertFn: func(t *testing.T, result []*commonpb.Payload, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "key not found")
			},
		},
		"enabled flag false - return as-is": {
			cfg: &internal.Config{},
			payloads: []*commonpb.Payload{
				{
					Metadata: map[string][]byte{
						converter.MetadataEncoding: []byte("nuon/s3payload"),
						"nuon/s3payload/enabled":   []byte("false"),
					},
					Data: []byte("temporal/blbabc123"),
				},
			},
			blobService: newMockBlobService(),
			assertFn: func(t *testing.T, result []*commonpb.Payload, err error) {
				require.NoError(t, err)
				require.Len(t, result, 1)

				// Should return unchanged
				assert.Equal(t, []byte("temporal/blbabc123"), result[0].Data)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			dc := &dataConverter{
				cfg:     test.cfg,
				l:       logger,
				blobSvc: test.blobService,
				mw:      mw,
			}

			result, err := dc.Decode(test.payloads)
			test.assertFn(t, result, err)
		})
	}
}

func TestDataConverter_RoundTrip(t *testing.T) {
	logger := zap.NewNop()
	mw := &mockMetricsWriter{}
	blobService := newMockBlobService()

	originalData := make([]byte, 2000) // Large payload
	for i := range originalData {
		originalData[i] = byte('A' + (i % 26))
	}

	tests := map[string]struct {
		payloads []*commonpb.Payload
		cfg      *internal.Config
		assertFn func(*testing.T, []*commonpb.Payload)
	}{
		"round trip - large payload": {
			cfg: &internal.Config{
				TemporalDataConverterLargePayloadSize: 1000,
			},
			payloads: []*commonpb.Payload{
				{
					Metadata: map[string][]byte{
						converter.MetadataEncoding: []byte("json"),
					},
					Data: originalData,
				},
			},
			assertFn: func(t *testing.T, decoded []*commonpb.Payload) {
				require.Len(t, decoded, 1)

				// Should have original data
				assert.Equal(t, originalData, decoded[0].Data)

				// Should have original encoding restored
				assert.Equal(t, []byte("json"), decoded[0].Metadata[converter.MetadataEncoding])
			},
		},
		"round trip - multiple payloads": {
			cfg: &internal.Config{
				TemporalDataConverterLargePayloadSize: 1000,
			},
			payloads: []*commonpb.Payload{
				{
					Metadata: map[string][]byte{converter.MetadataEncoding: []byte("json")},
					Data:     []byte("small"),
				},
				{
					Metadata: map[string][]byte{converter.MetadataEncoding: []byte("json")},
					Data:     originalData,
				},
			},
			assertFn: func(t *testing.T, decoded []*commonpb.Payload) {
				require.Len(t, decoded, 2)

				// First should be unchanged
				assert.Equal(t, []byte("small"), decoded[0].Data)

				// Second should have original data restored
				assert.Equal(t, originalData, decoded[1].Data)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			dc := &dataConverter{
				cfg:     test.cfg,
				l:       logger,
				blobSvc: blobService,
				mw:      mw,
			}

			// Encode
			encoded, err := dc.Encode(test.payloads)
			require.NoError(t, err)

			// Decode
			decoded, err := dc.Decode(encoded)
			require.NoError(t, err)

			// Assert
			test.assertFn(t, decoded)
		})
	}
}
