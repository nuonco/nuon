package log

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func TestLogStreamTransportPreservesRecordsAndRetries(t *testing.T) {
	payload := plogotlp.NewExportRequest()
	for _, message := range []string{"planning", "executing"} {
		r := payload.Logs().ResourceLogs().AppendEmpty()
		r.SetSchemaUrl("https://example.com/environment-schema")
		r.Resource().Attributes().PutStr("unexpected", "environment")
		r.Resource().SetDroppedAttributesCount(3)
		for _, scope := range []string{"workflow", "activity"} {
			s := r.ScopeLogs().AppendEmpty()
			s.Scope().SetName(scope)
			s.Scope().SetVersion("test-version")
			s.SetSchemaUrl("https://example.com/log-schema")
			l := s.LogRecords().AppendEmpty()
			l.Body().SetStr(message)
			l.Attributes().PutStr("phase", message)
			l.SetTraceID(pcommon.TraceID{1, 2, 3})
			l.SetSpanID(pcommon.SpanID{4, 5, 6})
			l.SetTimestamp(123)
			l.SetObservedTimestamp(456)
		}
	}
	original, err := payload.MarshalProto()
	require.NoError(t, err)
	req, err := http.NewRequest(http.MethodPost, "https://runner.example.com/v1/log-streams/test/logs", bytes.NewReader(original))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer stream-token")
	req.Header.Set("Content-Type", "application/x-protobuf")
	var sent [][]byte
	transport := &logStreamTransport{
		resource: getResource("stream-test", map[string]string{"operation": "build"}),
		next: roundTripperFunc(func(out *http.Request) (*http.Response, error) {
			require.NotSame(t, req, out)
			require.Equal(t, req.Method, out.Method)
			require.Equal(t, req.URL, out.URL)
			require.Equal(t, req.Header, out.Header)
			body, err := io.ReadAll(out.Body)
			require.NoError(t, err)
			require.NoError(t, out.Body.Close())
			require.EqualValues(t, len(body), out.ContentLength)
			replay, err := out.GetBody()
			require.NoError(t, err)
			replayed, err := io.ReadAll(replay)
			require.NoError(t, err)
			require.NoError(t, replay.Close())
			require.Equal(t, body, replayed)
			sent = append(sent, body)
			return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
		}),
	}
	for i := 0; i < 2; i++ {
		req.Body, err = req.GetBody()
		require.NoError(t, err)
		_, err = transport.RoundTrip(req)
		require.NoError(t, err)
		require.EqualValues(t, len(original), req.ContentLength)
	}
	require.Equal(t, sent[0], sent[1])
	decoded := plogotlp.NewExportRequest()
	require.NoError(t, decoded.UnmarshalProto(sent[0]))
	resources := decoded.Logs().ResourceLogs()
	require.Equal(t, 2, resources.Len())
	for i := 0; i < resources.Len(); i++ {
		r := resources.At(i)
		require.Empty(t, r.SchemaUrl())
		require.Zero(t, r.Resource().DroppedAttributesCount())
		require.Equal(t, map[string]any{"service.name": "api", "log_stream.id": "stream-test", "operation": "build"}, r.Resource().Attributes().AsRaw())
		originalResource := payload.Logs().ResourceLogs().At(i)
		originalResource.Resource().CopyTo(r.Resource())
		r.SetSchemaUrl(originalResource.SchemaUrl())
	}
	restored, err := decoded.MarshalProto()
	require.NoError(t, err)
	require.Equal(t, original, restored)
}
