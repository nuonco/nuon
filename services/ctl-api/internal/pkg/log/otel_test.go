package log

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func TestLogStreamIgnoresOTELEnvironment(t *testing.T) {
	for _, prefix := range []string{"OTEL_EXPORTER_OTLP_", "OTEL_EXPORTER_OTLP_LOGS_"} {
		t.Run(prefix, func(t *testing.T) {
			type request struct {
				path, authorization, encoding, extraHeader string
				body                                       []byte
			}
			requests := make(chan request, 16)
			receiver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				select {
				case <-r.Context().Done():
					return
				case <-time.After(20 * time.Millisecond):
				}
				requests <- request{r.URL.Path, r.Header.Get("Authorization"), r.Header.Get("Content-Encoding"), r.Header.Get("X-Collector"), body}
				w.Header().Set("Content-Type", "application/x-protobuf")
			}))
			t.Cleanup(receiver.Close)

			for _, p := range []string{"OTEL_EXPORTER_OTLP_", "OTEL_EXPORTER_OTLP_LOGS_"} {
				for _, key := range []string{"ENDPOINT", "HEADERS", "INSECURE", "PROTOCOL", "CERTIFICATE", "CLIENT_CERTIFICATE", "CLIENT_KEY", "TIMEOUT", "COMPRESSION"} {
					t.Setenv(p+key, "")
				}
			}
			for key, value := range map[string]string{
				"ENDPOINT":           "https://collector.invalid/other",
				"HEADERS":            "Authorization=collector-token,X-Collector=unexpected",
				"INSECURE":           "false",
				"PROTOCOL":           "grpc",
				"CERTIFICATE":        filepath.Join(t.TempDir(), "missing-ca.pem"),
				"CLIENT_CERTIFICATE": filepath.Join(t.TempDir(), "missing-cert.pem"),
				"CLIENT_KEY":         filepath.Join(t.TempDir(), "missing-key.pem"),
				"TIMEOUT":            "1",
				"COMPRESSION":        "gzip",
			} {
				t.Setenv(prefix+key, value)
			}
			for key, value := range map[string]string{
				"OTEL_BLRP_MAX_QUEUE_SIZE":                    "1",
				"OTEL_BLRP_MAX_EXPORT_BATCH_SIZE":             "1",
				"OTEL_BLRP_SCHEDULE_DELAY":                    "3600000",
				"OTEL_BLRP_EXPORT_TIMEOUT":                    "1",
				"OTEL_LOGRECORD_ATTRIBUTE_COUNT_LIMIT":        "1",
				"OTEL_LOGRECORD_ATTRIBUTE_VALUE_LENGTH_LIMIT": "1",
				"OTEL_SERVICE_NAME":                           "collector-service",
				"OTEL_RESOURCE_ATTRIBUTES":                    "service.name=other,log_stream.id=other,unexpected=value",
				"OTEL_SDK_DISABLED":                           "true",
				"OTEL_LOGS_EXPORTER":                          "none",
			} {
				t.Setenv(key, value)
			}

			stream := &app.LogStream{ID: "stream-test", RunnerAPIURL: receiver.URL, WriteToken: "stream-token"}
			provider, err := NewOTELProvider(stream)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
			logger, err := NewLogStreamLogger(stream, provider, zap.NewNop())
			require.NoError(t, err)
			for i := 0; i < 8; i++ {
				logger.Info("planning deployment", zap.Int("sequence", i), zap.String("phase", "planning"))
			}

			select {
			case req := <-requests:
				require.Equal(t, "/v1/log-streams/stream-test/logs", req.path)
				require.Equal(t, "Bearer stream-token", req.authorization)
				require.Empty(t, req.extraHeader)
				require.Empty(t, req.encoding)
				payload := plogotlp.NewExportRequest()
				require.NoError(t, payload.UnmarshalProto(req.body))
				resources := payload.Logs().ResourceLogs()
				require.Equal(t, 1, resources.Len())
				require.Equal(t, map[string]any{"service.name": "api", "log_stream.id": "stream-test"}, resources.At(0).Resource().Attributes().AsRaw())
				records := resources.At(0).ScopeLogs().At(0).LogRecords()
				require.Equal(t, 8, records.Len())
				for i := 0; i < records.Len(); i++ {
					record := records.At(i)
					require.Equal(t, "planning deployment", record.Body().Str())
					require.Equal(t, map[string]any{"sequence": int64(i), "phase": "planning"}, record.Attributes().AsRaw())
					require.Zero(t, record.DroppedAttributesCount())
				}
			case <-time.After(5 * time.Second):
				t.Fatal("workflow logs were not exported with their own batching and transport settings")
			}
			require.Equal(t, "https://collector.invalid/other", os.Getenv(prefix+"ENDPOINT"))
			require.Equal(t, "1", os.Getenv("OTEL_BLRP_MAX_QUEUE_SIZE"))
		})
	}
}

func TestLogStreamTLSIgnoresOTELTrust(t *testing.T) {
	receiver := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-protobuf")
	}))
	t.Cleanup(receiver.Close)
	certPath := filepath.Join(t.TempDir(), "collector-ca.pem")
	require.NoError(t, os.WriteFile(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: receiver.Certificate().Raw}), 0600))
	for _, prefix := range []string{"OTEL_EXPORTER_OTLP_", "OTEL_EXPORTER_OTLP_LOGS_"} {
		t.Run(prefix, func(t *testing.T) {
			exportErrors := make(chan error, 4)
			previousHandler := otel.GetErrorHandler()
			otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) { exportErrors <- err }))
			t.Cleanup(func() { otel.SetErrorHandler(previousHandler) })
			t.Setenv(prefix+"CERTIFICATE", certPath)
			t.Setenv(prefix+"INSECURE", "true")
			t.Setenv("OTEL_BLRP_SCHEDULE_DELAY", "3600000")
			provider, err := NewOTELProvider(&app.LogStream{ID: "stream-test", RunnerAPIURL: receiver.URL, WriteToken: "stream-token"})
			require.NoError(t, err)
			t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
			logger, err := NewLogStreamLogger(nil, provider, zap.NewNop())
			require.NoError(t, err)
			logger.Info("planning deployment")
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			require.NoError(t, provider.ForceFlush(ctx))
			select {
			case err := <-exportErrors:
				var unknownAuthority x509.UnknownAuthorityError
				require.ErrorAs(t, err, &unknownAuthority)
			case <-ctx.Done():
				t.Fatal("OTEL certificate settings changed log-stream TLS trust")
			}
		})
	}
}

func TestLogStreamRejectsInvalidEndpoint(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://collector.invalid")
	for _, endpoint := range []string{"", "http://%", "ftp://example.com"} {
		t.Run(endpoint, func(t *testing.T) {
			provider, err := NewOTELProvider(&app.LogStream{ID: "stream-test", RunnerAPIURL: endpoint})
			if provider != nil {
				t.Cleanup(func() { _ = provider.Shutdown(context.Background()) })
			}
			require.Error(t, err)
			require.Nil(t, provider)
		})
	}
}
