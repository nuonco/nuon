package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
	otlplogs "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

// Run with the two built binaries; no cluster or cloud account is contacted.
func TestAgentRuntime(t *testing.T) {
	binary, collector := os.Getenv("NUON_TEST_TELEMETRY_AGENT"), os.Getenv("NUON_TEST_OTELCOL")
	if binary == "" || collector == "" {
		t.Skip("set NUON_TEST_TELEMETRY_AGENT and NUON_TEST_OTELCOL to built binaries")
	}
	if runtime.GOOS != "linux" {
		t.Skip("SSL_CERT_DIR system trust is Linux-specific; use the Dockerfile test target")
	}
	var reserved []net.Listener
	newAddress := func() string {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		reserved = append(reserved, listener)
		return listener.Addr().String()
	}
	grpcAddress, httpAddress, childHealthAddress, health := newAddress(), newAddress(), newAddress(), newAddress()
	t.Cleanup(func() {
		for _, listener := range reserved {
			listener.Close()
		}
	})
	type received struct {
		relay, signal, token string
		attributes           map[string]any
	}
	forwarded := make(chan received, 32)
	var unavailable, rejectIssuance atomic.Bool
	var failures, settingsRequests atomic.Int64
	newRelay := func(name string, certificate tls.Certificate) *httptest.Server {
		server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if unavailable.Load() {
				failures.Add(1)
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			var reader io.Reader = r.Body
			if r.Header.Get("Content-Encoding") == "gzip" {
				compressed, err := gzip.NewReader(reader)
				if err != nil {
					t.Error(err)
					w.WriteHeader(400)
					return
				}
				defer compressed.Close()
				reader = compressed
			}
			contents, err := io.ReadAll(reader)
			if err != nil {
				t.Error(err)
				return
			}
			var attributes map[string]any
			switch r.URL.Path {
			case "/v1/logs":
				logs, err := (&plog.ProtoUnmarshaler{}).UnmarshalLogs(contents)
				if err != nil || logs.LogRecordCount() != 1 {
					t.Errorf("logs: count=%d error=%v", logs.LogRecordCount(), err)
					return
				}
				attributes = logs.ResourceLogs().At(0).Resource().Attributes().AsRaw()
			case "/v1/metrics":
				metrics, err := (&pmetric.ProtoUnmarshaler{}).UnmarshalMetrics(contents)
				if err != nil || metrics.DataPointCount() != 1 {
					t.Errorf("metrics: count=%d error=%v", metrics.DataPointCount(), err)
					return
				}
				attributes = metrics.ResourceMetrics().At(0).Resource().Attributes().AsRaw()
			case "/v1/traces":
				traces, err := (&ptrace.ProtoUnmarshaler{}).UnmarshalTraces(contents)
				if err != nil || traces.SpanCount() != 1 {
					t.Errorf("traces: count=%d error=%v", traces.SpanCount(), err)
					return
				}
				attributes = traces.ResourceSpans().At(0).Resource().Attributes().AsRaw()
			default:
				t.Errorf("unexpected relay path: %s", r.URL.Path)
			}
			forwarded <- received{name, r.URL.Path, r.Header.Get("Authorization"), attributes}
			w.Header().Set("Content-Type", "application/x-protobuf")
			w.WriteHeader(http.StatusOK)
		}))
		server.TLS = &tls.Config{Certificates: []tls.Certificate{certificate}}
		server.StartTLS()
		t.Cleanup(server.Close)
		return server
	}
	// Relay A is trusted only through SSL_CERT_FILE (the Collector's ca_file) and relay B only
	// through SSL_CERT_DIR, so reaching B proves the system pool survives a custom CA file.
	caA, certificateA := newTestCA(t, "relay-a")
	caB, certificateB := newTestCA(t, "relay-b")
	relayA, relayB := newRelay("a", certificateA), newRelay("b", certificateB)
	var settings atomic.Pointer[models.ServiceInstallTelemetryCollectorSettings]
	settings.Store(&models.ServiceInstallTelemetryCollectorSettings{RelayEndpoint: relayA.URL})
	var credential atomic.Value
	credential.Store("bootstrap-before")
	issued := make(chan string, 64)
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Header.Get("Authorization") != "Bearer "+credential.Load().(string) {
			w.WriteHeader(401)
			return
		}
		switch r.URL.Path {
		case "/v1/installs/inl_test/telemetry/collector-settings":
			settingsRequests.Add(1)
			json.NewEncoder(w).Encode(settings.Load())
		case "/v1/installs/inl_test/telemetry/access-token":
			if r.URL.Query().Get("relay_endpoint") != settings.Load().RelayEndpoint || rejectIssuance.Load() {
				w.WriteHeader(503)
				return
			}
			token := "jwt-" + credential.Load().(string)
			issued <- token
			json.NewEncoder(w).Encode(map[string]any{"access_token": token, "token_type": "Bearer", "expires_in": 4})
		default:
			t.Errorf("unexpected control-plane request: %s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer api.Close()
	directory := t.TempDir()
	caFile, caDirectory := filepath.Join(directory, "ca-a.pem"), filepath.Join(directory, "certs")
	if err := os.Mkdir(caDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	for path, ca := range map[string]*x509.Certificate{caFile: caA, filepath.Join(caDirectory, "ca-b.pem"): caB} {
		if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: ca.Raw}), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	trusts := func(roots []*x509.Certificate, relay *httptest.Server) bool {
		pool := x509.NewCertPool()
		for _, root := range roots {
			pool.AddCert(root)
		}
		_, err := relay.Certificate().Verify(x509.VerifyOptions{Roots: pool, DNSName: "127.0.0.1"})
		return err == nil
	}
	if !trusts([]*x509.Certificate{caA}, relayA) || trusts([]*x509.Certificate{caA}, relayB) || !trusts([]*x509.Certificate{caB}, relayB) || trusts([]*x509.Certificate{caB}, relayA) {
		t.Fatal("relay certificates must be trusted only by their own CA")
	}
	t.Setenv("SSL_CERT_FILE", caFile)
	t.Setenv("SSL_CERT_DIR", caDirectory)
	tokenFile := filepath.Join(directory, "bootstrap")
	if err := os.WriteFile(tokenFile, []byte("bootstrap-before\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	logFile, err := os.Create(filepath.Join(directory, "agent.log"))
	if err != nil {
		t.Fatal(err)
	}
	defer logFile.Close()
	var cmd *exec.Cmd
	start := func() {
		for _, listener := range reserved {
			listener.Close()
		}
		cmd = exec.Command(binary, "--api-url="+api.URL, "--allow-insecure-api", "--install-id=inl_test", "--token-file="+tokenFile,
			"--collector-binary="+collector, "--data-dir="+filepath.Join(directory, "state"), "--health-address="+health,
			"--otlp-grpc-address="+grpcAddress, "--otlp-http-address="+httpAddress, "--collector-health-address="+childHealthAddress)
		cmd.Stdout, cmd.Stderr = logFile, logFile
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
	}
	stop := func() {
		if cmd == nil {
			return
		}
		cmd.Process.Signal(syscall.SIGTERM)
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("agent shutdown: %v", err)
			}
		case <-time.After(10 * time.Second):
			cmd.Process.Kill()
			<-done
			t.Error("agent shutdown timed out")
		}
		cmd = nil
	}
	t.Cleanup(func() {
		stop()
		if t.Failed() {
			contents, _ := os.ReadFile(logFile.Name())
			t.Log(string(contents))
		}
	})
	httpClient := &http.Client{Timeout: time.Second}
	status := func(path string) int {
		response, err := httpClient.Get("http://" + health + path)
		if err != nil {
			return 0
		}
		response.Body.Close()
		return response.StatusCode
	}
	wait := func(description string, predicate func() bool) {
		t.Helper()
		deadline := time.Now().Add(22 * time.Second)
		for time.Now().Before(deadline) {
			if predicate() {
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
		t.Fatalf("timed out: %s", description)
	}
	waitForwarded := func(relay, signal, name, token string) {
		t.Helper()
		select {
		case result := <-forwarded:
			if result.relay != relay || result.signal != signal || result.token != "Bearer "+token || result.attributes["service.name"] != "producer" || result.attributes["nuon.install.name"] != name {
				t.Fatalf("unexpected forwarding: %#v", result)
			}
			if result.attributes["nuon.install.labels.literal"] != "${env:NOT_A_SECRET}" {
				t.Fatalf("metadata expanded: %#v", result.attributes)
			}
			if _, stale := result.attributes["nuon.install.labels.deleted"]; stale {
				t.Fatalf("stale label preserved: %#v", result.attributes)
			}
		case <-time.After(10 * time.Second):
			t.Fatal("telemetry was not forwarded")
		}
	}
	resource := `{"attributes":[{"key":"service.name","value":{"stringValue":"producer"}},{"key":"nuon.install.labels.deleted","value":{"stringValue":"stale"}}]}`
	payloads := map[string]string{
		"logs":    `{"resourceLogs":[{"resource":` + resource + `,"scopeLogs":[{"logRecords":[{"body":{"stringValue":"test"}}]}]}]}`,
		"metrics": `{"resourceMetrics":[{"resource":` + resource + `,"scopeMetrics":[{"metrics":[{"name":"requests","gauge":{"dataPoints":[{"asDouble":13.3}]}}]}]}]}`,
		"traces":  `{"resourceSpans":[{"resource":` + resource + `,"scopeSpans":[{"spans":[{"traceId":"0123456789abcdef0123456789abcdef","spanId":"0123456789abcdef","name":"test"}]}]}]}`,
	}
	send := func(signal string) {
		t.Helper()
		request, _ := http.NewRequest(http.MethodPost, "http://"+httpAddress+"/v1/"+signal, strings.NewReader(payloads[signal]))
		request.Header.Set("Content-Type", "application/json")
		response, err := httpClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		var result struct {
			PartialSuccess map[string]any `json:"partialSuccess"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			t.Fatalf("invalid OTLP response: %s: %v", body, err)
		}
		if response.StatusCode != 200 || len(result.PartialSuccess) != 0 {
			t.Fatalf("OTLP export response: status=%d body=%s", response.StatusCode, body)
		}
	}
	start()
	wait("disabled settings loaded", func() bool { return status("/readyz") == 200 })
	wait("settings polling while disabled", func() bool { return settingsRequests.Load() >= 2 })
	if len(issued) != 0 {
		t.Fatal("disabled collector requested a JWT")
	}
	attributes := map[string]string{"nuon.install.name": "before", "nuon.install.labels.literal": "${env:NOT_A_SECRET}"}
	settings.Store(&models.ServiceInstallTelemetryCollectorSettings{Enabled: true, RelayEndpoint: relayA.URL, ResourceAttributes: attributes})
	wait("enabled OTLP listener", func() bool {
		connection, err := net.DialTimeout("tcp", httpAddress, time.Second)
		if err != nil {
			return false
		}
		connection.Close()
		return true
	})
	for _, signal := range []string{"logs", "metrics", "traces"} {
		send(signal)
		waitForwarded("a", "/v1/"+signal, "before", "jwt-bootstrap-before")
	}
	connection, err := grpc.NewClient(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	request := &otlplogs.ExportLogsServiceRequest{}
	if err := protojson.Unmarshal([]byte(payloads["logs"]), request); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	_, err = otlplogs.NewLogsServiceClient(connection).Export(ctx, request)
	cancel()
	connection.Close()
	if err != nil {
		t.Fatal(err)
	}
	waitForwarded("a", "/v1/logs", "before", "jwt-bootstrap-before")
	credential.Store("bootstrap-after")
	if err := os.WriteFile(tokenFile+".next", []byte("bootstrap-after\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(tokenFile+".next", tokenFile); err != nil {
		t.Fatal(err)
	}
	wait("JWT renewal with rotated bootstrap", func() bool {
		select {
		case token := <-issued:
			return token == "jwt-bootstrap-after"
		default:
			return false
		}
	})
	wait("renewed JWT installed", func() bool {
		contents, _ := os.ReadFile(filepath.Join(directory, "state", "vendor-auth", "access-token"))
		return string(contents) == "jwt-bootstrap-after"
	})
	send("logs")
	waitForwarded("a", "/v1/logs", "before", "jwt-bootstrap-after")
	unavailable.Store(true)
	send("logs")
	wait("relay outage reached", func() bool { return failures.Load() > 0 })
	stop()
	unavailable.Store(false)
	start()
	waitForwarded("a", "/v1/logs", "before", "jwt-bootstrap-after")
	wait("restarted agent healthy", func() bool { return status("/readyz") == 200 })
	rejectIssuance.Store(true)
	settings.Store(&models.ServiceInstallTelemetryCollectorSettings{Enabled: true, RelayEndpoint: relayB.URL, ResourceAttributes: map[string]string{
		"nuon.install.name": "after", "nuon.install.labels.literal": "${env:NOT_A_SECRET}",
	}})
	wait("failed endpoint switch stops old collector", func() bool { return status("/readyz") == 503 })
	wait("old destination listener stopped", func() bool {
		connection, err := net.DialTimeout("tcp", httpAddress, time.Second)
		if err == nil {
			connection.Close()
		}
		return err != nil
	})
	rejectIssuance.Store(false)
	wait("new destination starts", func() bool { return status("/readyz") == 200 })
	send("metrics")
	waitForwarded("b", "/v1/metrics", "after", "jwt-bootstrap-after")
	settings.Store(&models.ServiceInstallTelemetryCollectorSettings{RelayEndpoint: relayB.URL})
	wait("disabled credentials removed", func() bool {
		_, err := os.Stat(filepath.Join(directory, "state", "vendor-auth"))
		return os.IsNotExist(err)
	})
	if status("/readyz") != 200 || status("/livez") != 200 {
		t.Fatal("disabled agent is unhealthy")
	}
	stop()
	contents, _ := os.ReadFile(logFile.Name())
	if bytes.Contains(contents, []byte("bootstrap-before")) || bytes.Contains(contents, []byte("bootstrap-after")) {
		t.Fatal("credentials leaked into logs")
	}
	t.Log("CA file plus SSL_CERT_DIR trust, disabled polling, HTTP/gRPC ingestion, all signals, literal enrichment, Secret rotation, JWT renewal, durable restart, failed destination switch, enable/disable, and shutdown passed")
}

// newTestCA returns a distinct root and a loopback relay certificate signed by it.
func newTestCA(t *testing.T, name string) (*x509.Certificate, tls.Certificate) {
	t.Helper()
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: name + " test CA"},
		NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour),
		IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageCertSign,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	ca, err := x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatal(err)
	}
	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	leafTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2), Subject: pkix.Name{CommonName: name},
		NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour),
		KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, leafTemplate, ca, &leafKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	return ca, tls.Certificate{Certificate: [][]byte{leafDER}, PrivateKey: leafKey}
}
