package strace

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"

	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"
)

const (
	defaultOTLPTracesEndpointTmpl string = "%s/v1/runners/%s/traces"
)

// jsonHTTPClient implements otlptrace.Client by POSTing JSON-encoded OTLP
// trace export requests to the runner traces ingest endpoint.
//
// We deliberately do NOT use otlptracehttp here: the existing CTL-API trace
// ingest endpoint accepts JSON only (it unmarshals the request body via
// ptraceotlp.NewExportRequest().UnmarshalJSON), and we want to match the
// established slog/otel.go pattern of POSTing OTLP JSON via the runner API.
//
// NOTE: the runner SDK's WriteOTELTraces is currently a no-op stub, so we
// bypass it the same way the log exporter bypasses WriteOTELLogs in favor of
// otlploghttp posting directly to /v1/log-streams/{id}/logs.
type jsonHTTPClient struct {
	endpoint string
	token    string
	http     *http.Client

	mu     sync.Mutex
	closed bool
}

var _ otlptrace.Client = (*jsonHTTPClient)(nil)

func newJSONHTTPClient(apiURL, runnerID, token string) *jsonHTTPClient {
	return &jsonHTTPClient{
		endpoint: fmt.Sprintf(defaultOTLPTracesEndpointTmpl, apiURL, runnerID),
		token:    token,
		http:     http.DefaultClient,
	}
}

func (c *jsonHTTPClient) Start(ctx context.Context) error { return nil }

func (c *jsonHTTPClient) Stop(ctx context.Context) error {
	c.mu.Lock()
	c.closed = true
	c.mu.Unlock()
	return nil
}

func (c *jsonHTTPClient) UploadTraces(ctx context.Context, protoSpans []*tracepb.ResourceSpans) error {
	c.mu.Lock()
	closed := c.closed
	c.mu.Unlock()
	if closed {
		return nil
	}
	if len(protoSpans) == 0 {
		return nil
	}

	// Encode via ptraceotlp so trace_id / span_id land as hex strings, which
	// is what the OTLP/JSON spec (and the ctl-api ingest endpoint, which
	// parses with ptraceotlp.NewExportRequest().UnmarshalJSON) requires.
	// google.golang.org/protobuf/encoding/protojson would emit them as base64
	// (the proto bytes default) and the server would 500 with
	// "ID.UnmarshalJSONIter: length mismatch".
	protoBytes, err := proto.Marshal(&coltracepb.ExportTraceServiceRequest{ResourceSpans: protoSpans})
	if err != nil {
		return fmt.Errorf("strace: marshal export request to proto: %w", err)
	}
	exportReq := ptraceotlp.NewExportRequest()
	if err := exportReq.UnmarshalProto(protoBytes); err != nil {
		return fmt.Errorf("strace: convert proto to ptraceotlp: %w", err)
	}
	body, err := exportReq.MarshalJSON()
	if err != nil {
		return fmt.Errorf("strace: marshal export request to json: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("strace: build http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("strace: post traces: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("strace: traces ingest returned %d: %s", resp.StatusCode, string(respBody))
	}
	return nil
}
