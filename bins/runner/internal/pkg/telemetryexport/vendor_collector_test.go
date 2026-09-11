package telemetryexport

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"maps"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"gopkg.in/yaml.v3"
)

// Run with NUON_TEST_OTELCOL pointing to the built runner Collector.
func TestVendorCollectorEnrichesResourceAttributes(t *testing.T) {
	binary := os.Getenv("NUON_TEST_OTELCOL")
	if binary == "" {
		t.Skip("set NUON_TEST_OTELCOL to the built runner Collector binary")
	}
	t.Setenv("NUON_ATTRIBUTE_TEST", "must-not-expand")
	attributes := map[string]string{
		"nuon.org.name": "acme", "nuon.app.name": "payments", "nuon.install.name": "production-eu",
		"nuon.install.labels.tier":                               "enterprise",
		"nuon.install.labels.example.com/team.name":              "platform",
		"nuon.install.labels.literal-${env:NUON_ATTRIBUTE_TEST}": "$ ${env:NUON_ATTRIBUTE_TEST} $$\nnext line",
		"nuon.install.labels.empty":                              "", "nuon.install.labels.number": "001", "nuon.install.labels.boolean": "true",
	}
	type exported struct {
		path string
		body []byte
	}
	outputs := make(chan exported, 10)
	backend := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("missing vendor bearer token")
		}
		outputs <- exported{path: r.URL.Path, body: body}
		w.Header().Set("Content-Type", "application/x-protobuf")
	}))
	defer backend.Close()
	contents, err := vendorCollectorConfig(backend.URL, attributes)
	require.NoError(t, err)
	other, err := vendorCollectorConfig(backend.URL, maps.Clone(attributes))
	require.NoError(t, err)
	require.Equal(t, contents, other, "attribute action order must be deterministic")
	var document map[string]any
	require.NoError(t, yaml.Unmarshal(contents, &document))

	// Isolate network listeners, storage, and credentials from the host runner.
	address := func() string {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		addr := listener.Addr().String()
		require.NoError(t, listener.Close())
		return addr
	}
	otlpAddress, healthAddress := address(), address()
	document["receivers"].(map[string]any)["otlp"] = map[string]any{"protocols": map[string]any{
		"http": map[string]any{"endpoint": otlpAddress},
	}}
	extensions := document["extensions"].(map[string]any)
	extensions["health_check"] = map[string]any{"endpoint": healthAddress}
	extensions[vendorFileStorageExtensionID] = fileStorageConfig(t.TempDir())
	tokenPath := filepath.Join(t.TempDir(), "token")
	require.NoError(t, os.WriteFile(tokenPath, []byte("test-token"), 0o600))
	extensions[vendorBearerAuthExtensionID] = map[string]any{"filename": tokenPath}
	exporter := document["exporters"].(map[string]any)["otlp_http/vendor"].(map[string]any)
	exporter["tls"] = map[string]any{"insecure_skip_verify": true}
	exporter["compression"] = "none"
	document["service"].(map[string]any)["telemetry"] = map[string]any{"metrics": map[string]any{"level": "none"}}
	contents, err = yaml.Marshal(document)
	require.NoError(t, err)
	configPath := filepath.Join(t.TempDir(), "collector.yaml")
	require.NoError(t, os.WriteFile(configPath, contents, 0o600))
	logPath := filepath.Join(t.TempDir(), "collector.log")
	logFile, err := os.Create(logPath)
	require.NoError(t, err)
	defer logFile.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "--config", configPath)
	cmd.Cancel = func() error { return cmd.Process.Signal(syscall.SIGTERM) }
	cmd.WaitDelay = 3 * time.Second
	cmd.Stdout, cmd.Stderr = logFile, logFile
	require.NoError(t, cmd.Start())
	child := &childProcess{done: make(chan struct{})}
	go func() { _ = cmd.Wait(); close(child.done) }()
	defer func() {
		cancel()
		<-child.done
		if t.Failed() {
			logs, _ := os.ReadFile(logPath)
			t.Log(string(logs))
		}
	}()
	require.NoError(t, waitForCollector(ctx, child, "http://"+healthAddress))

	resourceAttributes := []map[string]any{
		{"key": "nuon.org.name", "value": map[string]any{"stringValue": "wrong-org"}},
		{"key": "nuon.app.name", "value": map[string]any{"stringValue": "wrong-app"}},
		{"key": "nuon.install.name", "value": map[string]any{"stringValue": "wrong-install"}},
		{"key": "nuon.install.labels.tier", "value": map[string]any{"stringValue": "wrong-tier"}},
		{"key": "nuon.install.labels.deleted", "value": map[string]any{"stringValue": "stale"}},
		{"key": "service.name", "value": map[string]any{"stringValue": "checkout"}},
		{"key": "cloud.region", "value": map[string]any{"stringValue": "eu-west-1"}},
		{"key": "nuon.org.id", "value": map[string]any{"stringValue": "producer-id"}},
		{"key": "nuonXinstallXlabelsXkeep", "value": map[string]any{"stringValue": "keep"}},
	}
	recordAttributes := []map[string]any{
		{"key": "nuon.install.name", "value": map[string]any{"stringValue": "record-name"}},
		{"key": "nuon.install.labels.deleted", "value": map[string]any{"stringValue": "record-label"}},
	}
	wantResource := map[string]any{
		"service.name": "checkout", "cloud.region": "eu-west-1", "nuon.org.id": "producer-id", "nuonXinstallXlabelsXkeep": "keep",
	}
	for key, value := range attributes {
		wantResource[key] = value
	}
	wantRecord := map[string]any{"nuon.install.name": "record-name", "nuon.install.labels.deleted": "record-label"}
	for _, signal := range []string{"logs", "metrics", "traces"} {
		t.Run(signal, func(t *testing.T) {
			var payload map[string]any
			switch signal {
			case "logs":
				payload = map[string]any{"resourceLogs": []any{map[string]any{
					"resource":  map[string]any{"attributes": resourceAttributes},
					"scopeLogs": []any{map[string]any{"logRecords": []any{map[string]any{"attributes": recordAttributes, "body": map[string]any{"stringValue": "hello"}}}}},
				}}}
			case "metrics":
				payload = map[string]any{"resourceMetrics": []any{map[string]any{
					"resource":     map[string]any{"attributes": resourceAttributes},
					"scopeMetrics": []any{map[string]any{"metrics": []any{map[string]any{"name": "requests", "gauge": map[string]any{"dataPoints": []any{map[string]any{"attributes": recordAttributes, "asInt": "7"}}}}}}},
				}}}
			case "traces":
				payload = map[string]any{"resourceSpans": []any{map[string]any{
					"resource":   map[string]any{"attributes": resourceAttributes},
					"scopeSpans": []any{map[string]any{"spans": []any{map[string]any{"name": "checkout", "traceId": "0102030405060708090a0b0c0d0e0f10", "spanId": "0102030405060708", "attributes": recordAttributes}}}},
				}}}
			}
			body, err := json.Marshal(payload)
			require.NoError(t, err)
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://"+otlpAddress+"/v1/"+signal, bytes.NewReader(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			response, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer response.Body.Close()
			require.Equal(t, http.StatusOK, response.StatusCode)
			select {
			case output := <-outputs:
				require.Equal(t, "/v1/"+signal, output.path)
				var resource, record pcommon.Map
				switch signal {
				case "logs":
					data := plogotlp.NewExportRequest()
					require.NoError(t, data.UnmarshalProto(output.body))
					require.Equal(t, 1, data.Logs().LogRecordCount())
					logs := data.Logs().ResourceLogs().At(0)
					resource, record = logs.Resource().Attributes(), logs.ScopeLogs().At(0).LogRecords().At(0).Attributes()
				case "metrics":
					data := pmetricotlp.NewExportRequest()
					require.NoError(t, data.UnmarshalProto(output.body))
					require.Equal(t, 1, data.Metrics().DataPointCount())
					metrics := data.Metrics().ResourceMetrics().At(0)
					resource, record = metrics.Resource().Attributes(), metrics.ScopeMetrics().At(0).Metrics().At(0).Gauge().DataPoints().At(0).Attributes()
				case "traces":
					data := ptraceotlp.NewExportRequest()
					require.NoError(t, data.UnmarshalProto(output.body))
					require.Equal(t, 1, data.Traces().SpanCount())
					spans := data.Traces().ResourceSpans().At(0)
					resource, record = spans.Resource().Attributes(), spans.ScopeSpans().At(0).Spans().At(0).Attributes()
				}
				require.Equal(t, wantResource, resource.AsRaw())
				require.Equal(t, wantRecord, record.AsRaw())
			case <-ctx.Done():
				t.Fatal("timed out waiting for exported telemetry")
			}
		})
	}
}
