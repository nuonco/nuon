package nuonpartialsuccessextension

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/proto"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestPartialRejectionsBecomePermanentHTTPFailures(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		rejected int64
		body     func(int64) ([]byte, error)
	}{
		{
			name: "logs", path: "/v1/logs", rejected: 2,
			body: func(rejected int64) ([]byte, error) {
				response := plogotlp.NewExportResponse()
				response.PartialSuccess().SetRejectedLogRecords(rejected)
				return response.MarshalProto()
			},
		},
		{
			name: "metrics", path: "/prefix/v1/metrics", rejected: 3,
			body: func(rejected int64) ([]byte, error) {
				response := pmetricotlp.NewExportResponse()
				response.PartialSuccess().SetRejectedDataPoints(rejected)
				return response.MarshalProto()
			},
		},
		{
			name: "traces", path: "/v1/traces", rejected: 4,
			body: func(rejected int64) ([]byte, error) {
				response := ptraceotlp.NewExportResponse()
				response.PartialSuccess().SetRejectedSpans(rejected)
				return response.MarshalProto()
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := test.body(test.rejected)
			require.NoError(t, err)
			guard := responseGuardRoundTripper{base: staticResponse(body, protobufContentType)}
			request, err := http.NewRequest(http.MethodPost, "https://vendor.example.com"+test.path, nil)
			require.NoError(t, err)

			response, err := guard.RoundTrip(request)

			require.NoError(t, err)
			require.Equal(t, http.StatusBadRequest, response.StatusCode)
			contents, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			var responseStatus status.Status
			require.NoError(t, proto.Unmarshal(contents, &responseStatus))
			require.Contains(t, responseStatus.Message, "signal="+test.name)
			require.Contains(t, responseStatus.Message, "rejected=")
		})
	}
}

func TestJSONPartialRejectionBecomesFailure(t *testing.T) {
	downstream := plogotlp.NewExportResponse()
	downstream.PartialSuccess().SetRejectedLogRecords(1)
	body, err := downstream.MarshalJSON()
	require.NoError(t, err)
	guard := responseGuardRoundTripper{base: staticResponse(body, jsonContentType)}
	request, err := http.NewRequest(http.MethodPost, "https://vendor.example.com/v1/logs", nil)
	require.NoError(t, err)

	response, err := guard.RoundTrip(request)

	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, response.StatusCode)
}

func TestFullSuccessResponseIsPreserved(t *testing.T) {
	downstream := plogotlp.NewExportResponse()
	downstream.PartialSuccess().SetErrorMessage("accepted with a warning")
	body, err := downstream.MarshalProto()
	require.NoError(t, err)
	guard := responseGuardRoundTripper{base: staticResponse(body, protobufContentType)}
	request, err := http.NewRequest(http.MethodPost, "https://vendor.example.com/v1/logs", nil)
	require.NoError(t, err)

	response, err := guard.RoundTrip(request)
	require.NoError(t, err)
	contents, err := io.ReadAll(response.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(t, body, contents)
}

func TestUninspectableResponseIsPreserved(t *testing.T) {
	tests := map[string]struct {
		body          []byte
		contentLength int64
	}{
		"malformed": {
			body:          []byte("not protobuf"),
			contentLength: int64(len("not protobuf")),
		},
		"oversized unknown length": {
			body:          bytes.Repeat([]byte("a"), maxResponseBodySize+1),
			contentLength: -1,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			guard := responseGuardRoundTripper{base: staticResponseWithLength(test.body, protobufContentType, test.contentLength)}
			request, err := http.NewRequest(http.MethodPost, "https://vendor.example.com/v1/logs", nil)
			require.NoError(t, err)

			response, err := guard.RoundTrip(request)
			require.NoError(t, err)
			contents, err := io.ReadAll(response.Body)
			require.NoError(t, err)

			require.Equal(t, http.StatusOK, response.StatusCode)
			require.Equal(t, test.body, contents)
		})
	}
}

func TestTransportErrorsPassThrough(t *testing.T) {
	expected := errors.New("transport failed")
	guard := responseGuardRoundTripper{base: roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, expected
	})}
	request, err := http.NewRequest(http.MethodPost, "https://vendor.example.com/v1/logs", nil)
	require.NoError(t, err)

	response, err := guard.RoundTrip(request)

	require.Nil(t, response)
	require.ErrorIs(t, err, expected)
}

func staticResponse(body []byte, contentType string) http.RoundTripper {
	return staticResponseWithLength(body, contentType, int64(len(body)))
}

func staticResponseWithLength(body []byte, contentType string, contentLength int64) http.RoundTripper {
	return roundTripperFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode:    http.StatusOK,
			Status:        "200 OK",
			Header:        http.Header{"Content-Type": []string{contentType}},
			Body:          io.NopCloser(bytes.NewReader(body)),
			ContentLength: contentLength,
			Request:       request,
		}, nil
	})
}
