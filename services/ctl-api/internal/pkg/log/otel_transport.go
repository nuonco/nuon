package log

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/otel/sdk/resource"
)

type logStreamTransport struct {
	next     http.RoundTripper
	resource *resource.Resource
}

func (t *logStreamTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body, err := io.ReadAll(req.Body)
	_ = req.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("read log stream export: %w", err)
	}
	payload := plogotlp.NewExportRequest()
	if err := payload.UnmarshalProto(body); err != nil {
		return nil, fmt.Errorf("decode log stream export: %w", err)
	}

	// The Logs SDK merges environment attributes even with an explicit resource.
	// Replace the wire resource without mutating the SDK's shared resource.
	resources := payload.Logs().ResourceLogs()
	for i := 0; i < resources.Len(); i++ {
		r := resources.At(i)
		r.SetSchemaUrl(t.resource.SchemaURL())
		r.Resource().SetDroppedAttributesCount(0)
		attrs := r.Resource().Attributes()
		attrs.Clear()
		for _, attr := range t.resource.Attributes() {
			attrs.PutStr(string(attr.Key), attr.Value.AsString())
		}
	}
	body, err = payload.MarshalProto()
	if err != nil {
		return nil, fmt.Errorf("encode log stream export: %w", err)
	}
	clone := req.Clone(req.Context())
	clone.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(body)), nil }
	clone.Body, _ = clone.GetBody()
	clone.ContentLength = int64(len(body))
	return t.next.RoundTrip(clone)
}
