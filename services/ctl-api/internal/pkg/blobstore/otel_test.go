package blobstore

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

type fakeService struct {
	downloadStream func(context.Context, string) (io.ReadCloser, error)
}

func (f *fakeService) Upload(context.Context, string, []byte) error { return nil }
func (f *fakeService) Download(context.Context, string) ([]byte, error) {
	return []byte("blob"), nil
}
func (f *fakeService) UploadStream(context.Context, string, io.Reader) (string, error) {
	return "sha256:checksum", nil
}
func (f *fakeService) DownloadStream(ctx context.Context, key string) (io.ReadCloser, error) {
	return f.downloadStream(ctx, key)
}
func (f *fakeService) GetMetadata(context.Context, string) (int64, string, error) {
	return 4, "application/octet-stream", nil
}

type trackingReadCloser struct {
	reader     io.Reader
	closeErr   error
	closeCalls int
}

func (r *trackingReadCloser) Read(p []byte) (int, error) { return r.reader.Read(p) }
func (r *trackingReadCloser) Close() error {
	r.closeCalls++
	return r.closeErr
}

func TestInstrumentedServiceOperations(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	svc := instrumentService(&fakeService{downloadStream: func(context.Context, string) (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewBufferString("blob")), nil
	}}, provider)

	require.NoError(t, svc.Upload(context.Background(), "sensitive-key", []byte("blob")))
	_, err := svc.Download(context.Background(), "sensitive-key")
	require.NoError(t, err)
	_, err = svc.UploadStream(context.Background(), "sensitive-key", bytes.NewBufferString("blob"))
	require.NoError(t, err)
	_, _, err = svc.GetMetadata(context.Background(), "sensitive-key")
	require.NoError(t, err)
	body, err := svc.DownloadStream(context.Background(), "sensitive-key")
	require.NoError(t, err)
	_, err = io.ReadAll(body)
	require.NoError(t, err)
	require.NoError(t, body.Close())

	counts := collectOperationCounts(t, reader)
	for _, operation := range []string{"write", "read", "write_stream", "metadata", "read_stream_open", "read_stream_body"} {
		require.EqualValues(t, 1, counts[operation+":success"])
	}
}

func TestInstrumentedStreamCompletionOutcomes(t *testing.T) {
	tests := map[string]struct {
		reader      io.Reader
		closeErr    error
		consume     bool
		wantOutcome string
	}{
		"read error":  {reader: io.MultiReader(bytes.NewBufferString("blob"), errorReader{}), consume: true, wantOutcome: "error"},
		"early close": {reader: bytes.NewBufferString("blob"), wantOutcome: "closed_early"},
		"close error": {reader: bytes.NewBufferString("blob"), closeErr: errors.New("close failed"), wantOutcome: "error"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			reader := sdkmetric.NewManualReader()
			provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
			t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
			underlying := &trackingReadCloser{reader: tt.reader, closeErr: tt.closeErr}
			svc := instrumentService(&fakeService{downloadStream: func(context.Context, string) (io.ReadCloser, error) {
				return underlying, nil
			}}, provider)

			body, err := svc.DownloadStream(context.Background(), "sensitive-key")
			require.NoError(t, err)
			if tt.consume {
				_, _ = io.ReadAll(body)
			}
			_ = body.Close()
			_ = body.Close()
			require.Equal(t, 2, underlying.closeCalls)

			counts := collectOperationCounts(t, reader)
			require.EqualValues(t, 1, counts["read_stream_open:success"])
			require.EqualValues(t, 1, counts["read_stream_body:"+tt.wantOutcome])
		})
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

func collectOperationCounts(t *testing.T, reader *sdkmetric.ManualReader) map[string]uint64 {
	t.Helper()
	var data metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &data))
	counts := make(map[string]uint64)
	for _, scope := range data.ScopeMetrics {
		for _, collected := range scope.Metrics {
			if collected.Name != "nuon.blobstore.operations" {
				continue
			}
			for _, point := range collected.Data.(metricdata.Sum[int64]).DataPoints {
				operation, _ := point.Attributes.Value(attribute.Key("operation"))
				outcome, _ := point.Attributes.Value(attribute.Key("outcome"))
				require.Equal(t, 2, point.Attributes.Len())
				counts[operation.AsString()+":"+outcome.AsString()] += uint64(point.Value)
			}
		}
	}
	return counts
}
