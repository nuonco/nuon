package telemetryexport

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestWaitForCollector(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	if err := waitForCollector(context.Background(), &Collector{done: make(chan struct{})}, server.URL); err != nil {
		t.Fatalf("waitForCollector() error = %v", err)
	}
}

func TestWaitForCollectorDetectsExitedProcess(t *testing.T) {
	done := make(chan struct{})
	close(done)
	if err := waitForCollector(context.Background(), &Collector{done: done}, "http://127.0.0.1:0"); err == nil {
		t.Fatal("waitForCollector() returned nil for exited process")
	}
}

func TestWaitForCollectorHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		cancel()
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()
	defer cancel()
	if err := waitForCollector(ctx, &Collector{done: make(chan struct{})}, server.URL); !errors.Is(err, context.Canceled) {
		t.Fatalf("health wait did not honor cancellation: %v", err)
	}
}

func TestCollectorReplacementAndCancellation(t *testing.T) {
	binary := os.Getenv("NUON_TEST_OTELCOL")
	if binary == "" {
		t.Skip("set NUON_TEST_OTELCOL to the built Collector binary")
	}
	configDirectory := t.TempDir()
	t.Setenv("TMPDIR", configDirectory)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	healthAddress := listener.Addr().String()
	listener.Close()
	cfg := DefaultConfig(t.TempDir())
	cfg.GRPCAddress, cfg.HTTPAddress, cfg.HealthAddress = "127.0.0.1:0", "127.0.0.1:0", healthAddress
	if err := os.MkdirAll(cfg.TokenDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfg.TokenDirectory, "access-token"), []byte("relay-token"), 0o600); err != nil {
		t.Fatal(err)
	}
	contents, err := CollectorConfig(cfg, "https://relay.example.com", nil)
	if err != nil {
		t.Fatal(err)
	}
	options := CollectorOptions{Binary: binary, Config: contents, HealthURL: "http://" + healthAddress + "/", Logger: zap.NewNop()}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	child, err := StartCollector(ctx, nil, options)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(child.Stop)
	checkHealthy := func() {
		t.Helper()
		select {
		case <-child.Done():
			t.Fatal("healthy Collector unexpectedly exited")
		default:
		}
		response, err := http.Get(options.HealthURL)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("Collector is not healthy: %s", response.Status)
		}
	}
	checkHealthy()
	missing := options
	missing.Binary = filepath.Join(t.TempDir(), "missing-collector")
	if retained, err := StartCollector(ctx, child, missing); err == nil || retained != child {
		t.Fatalf("preparation failure did not retain ownership of the running child: child=%v error=%v", retained, err)
	}
	checkHealthy()
	previous := child
	child, err = StartCollector(ctx, previous, options)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(child.Stop)
	select {
	case <-previous.Done():
	default:
		t.Fatal("replacement returned before previous child exited")
	}
	checkHealthy()
	cancel()
	select {
	case <-child.Done():
	case <-time.After(7 * time.Second):
		t.Fatal("Collector did not exit after cancellation")
	}
	child.Stop()
	if response, err := http.Get(options.HealthURL); err == nil {
		response.Body.Close()
		t.Fatal("Collector health listener survived shutdown")
	}
	invalid := options
	invalid.Config = []byte("not valid Collector YAML: [")
	if failed, err := StartCollector(context.Background(), nil, invalid); err == nil || failed != nil {
		t.Fatalf("failed launch retained a child: child=%v error=%v", failed, err)
	}
	files, err := os.ReadDir(configDirectory)
	if err != nil || len(files) != 0 {
		t.Fatalf("temporary configuration survived stopped or failed children: files=%v error=%v", files, err)
	}
}

func TestChildEnvironmentOmitsRunnerCredentials(t *testing.T) {
	t.Setenv("RUNNER_API_TOKEN", "runner-secret")
	t.Setenv("HTTPS_PROXY", "https://proxy.example.com")
	environment := childEnvironment([]string{"NUON_TELEMETRY_EXPORT_HEADER_0=customer-secret"})
	joined := strings.Join(environment, "\n")
	if strings.Contains(joined, "runner-secret") || strings.Contains(joined, "RUNNER_API_TOKEN") {
		t.Fatal("collector child inherited the runner API credential")
	}
	if !strings.Contains(joined, "HTTPS_PROXY=https://proxy.example.com") {
		t.Fatal("collector child did not inherit HTTPS proxy configuration")
	}
	if !strings.Contains(joined, "NUON_TELEMETRY_EXPORT_HEADER_0=customer-secret") {
		t.Fatal("collector child did not receive the customer header")
	}
}
