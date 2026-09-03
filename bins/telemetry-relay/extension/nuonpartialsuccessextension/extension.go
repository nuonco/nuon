package nuonpartialsuccessextension

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
	"go.opentelemetry.io/collector/extension/extensionmiddleware"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
)

const (
	protobufContentType = "application/x-protobuf"
	jsonContentType     = "application/json"
	maxResponseBodySize = 64 * 1024
	retryAfterSeconds   = "30"
)

type partialSuccessExtension struct {
	logger *zap.Logger
}

var (
	_ extension.Extension            = (*partialSuccessExtension)(nil)
	_ extensionmiddleware.HTTPClient = (*partialSuccessExtension)(nil)
	_ http.RoundTripper              = (*responseGuardRoundTripper)(nil)
)

func (*partialSuccessExtension) Start(context.Context, component.Host) error {
	return nil
}

func (*partialSuccessExtension) Shutdown(context.Context) error {
	return nil
}

func (e *partialSuccessExtension) GetHTTPRoundTripper(context.Context) (extensionmiddleware.WrapHTTPRoundTripperFunc, error) {
	return func(_ context.Context, base http.RoundTripper) (http.RoundTripper, error) {
		return &responseGuardRoundTripper{base: base, logger: e.logger}, nil
	}, nil
}

type responseGuardRoundTripper struct {
	base   http.RoundTripper
	logger *zap.Logger
}

func (r *responseGuardRoundTripper) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := r.base.RoundTrip(request)
	if err != nil || response == nil {
		return response, err
	}
	switch response.StatusCode {
	case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		response.StatusCode = http.StatusServiceUnavailable
		response.Status = fmt.Sprintf("%d %s", http.StatusServiceUnavailable, http.StatusText(http.StatusServiceUnavailable))
		response.Header = response.Header.Clone()
		if response.Header == nil {
			response.Header = make(http.Header)
		}
		response.Header.Set("Retry-After", retryAfterSeconds)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return response, nil
	}

	signal := signalFromPath(request.URL.Path)
	contentType := response.Header.Get("Content-Type")
	if signal == "" || (contentType != protobufContentType && contentType != jsonContentType) || response.Header.Get("Content-Encoding") != "" {
		return response, nil
	}

	contents, ok := boundedResponseBody(response)
	if !ok {
		return response, nil
	}
	rejected, ok := rejectedItems(signal, contentType, contents)
	if !ok || rejected == 0 {
		return response, nil
	}

	if r.logger != nil {
		r.logger.Warn("vendor backend partially rejected telemetry", zap.String("signal", signal), zap.Int64("rejected", rejected))
	}
	return partialRejectionResponse(response, signal, rejected), nil
}

func signalFromPath(path string) string {
	switch {
	case strings.HasSuffix(path, "/v1/logs"):
		return "logs"
	case strings.HasSuffix(path, "/v1/metrics"):
		return "metrics"
	case strings.HasSuffix(path, "/v1/traces"):
		return "traces"
	default:
		return ""
	}
}

func boundedResponseBody(response *http.Response) ([]byte, bool) {
	if response.Body == nil || response.ContentLength > maxResponseBodySize {
		return nil, false
	}

	original := response.Body
	contents, err := io.ReadAll(io.LimitReader(original, maxResponseBodySize+1))
	if err != nil || len(contents) > maxResponseBodySize {
		response.Body = &replayReadCloser{
			Reader: io.MultiReader(bytes.NewReader(contents), original),
			Closer: original,
		}
		return nil, false
	}
	_ = original.Close()
	response.Body = io.NopCloser(bytes.NewReader(contents))
	response.ContentLength = int64(len(contents))
	return contents, true
}

type replayReadCloser struct {
	io.Reader
	io.Closer
}

func rejectedItems(signal, contentType string, contents []byte) (int64, bool) {
	switch signal {
	case "logs":
		response := plogotlp.NewExportResponse()
		if !unmarshalResponse(contentType, contents, response.UnmarshalProto, response.UnmarshalJSON) {
			return 0, false
		}
		return response.PartialSuccess().RejectedLogRecords(), true
	case "metrics":
		response := pmetricotlp.NewExportResponse()
		if !unmarshalResponse(contentType, contents, response.UnmarshalProto, response.UnmarshalJSON) {
			return 0, false
		}
		return response.PartialSuccess().RejectedDataPoints(), true
	case "traces":
		response := ptraceotlp.NewExportResponse()
		if !unmarshalResponse(contentType, contents, response.UnmarshalProto, response.UnmarshalJSON) {
			return 0, false
		}
		return response.PartialSuccess().RejectedSpans(), true
	default:
		return 0, false
	}
}

func unmarshalResponse(contentType string, contents []byte, unmarshalProto, unmarshalJSON func([]byte) error) bool {
	var err error
	if contentType == protobufContentType {
		err = unmarshalProto(contents)
	} else {
		err = unmarshalJSON(contents)
	}
	return err == nil
}

func partialRejectionResponse(response *http.Response, signal string, rejected int64) *http.Response {
	message := fmt.Sprintf("downstream partial rejection: signal=%s rejected=%d", signal, rejected)
	contents, _ := proto.Marshal(&status.Status{
		Code:    int32(codes.InvalidArgument),
		Message: message,
	})
	response.StatusCode = http.StatusBadRequest
	response.Status = fmt.Sprintf("%d %s", http.StatusBadRequest, http.StatusText(http.StatusBadRequest))
	response.Header = response.Header.Clone()
	response.Header.Set("Content-Type", protobufContentType)
	response.Header.Set("Content-Length", fmt.Sprintf("%d", len(contents)))
	response.Header.Del("Content-Encoding")
	response.ContentLength = int64(len(contents))
	response.TransferEncoding = nil
	response.Body = io.NopCloser(bytes.NewReader(contents))
	return response
}
