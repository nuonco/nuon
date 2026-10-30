package http

import (
	"context"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestConfigFileOmitsTokenAndProxyAddsIt(t *testing.T) {
	const token = "runner-token"
	var sawToken string
	var sawPath string
	upstream := httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		sawToken = r.URL.Query().Get("token")
		sawPath = r.URL.Path
		w.WriteHeader(nethttp.StatusNoContent)
	}))
	t.Cleanup(upstream.Close)

	backend := &http{Config: &NuonWorkspaceConfig{
		APIEndpoint: upstream.URL,
		WorkspaceID: "tfw123",
		Token:       token,
		JobID:       "job123",
	}}
	if err := backend.Init(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = backend.Close() })

	byts, err := backend.ConfigFile(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(byts), token) {
		t.Fatalf("backend config contains the runner token: %s", byts)
	}
	if !strings.Contains(string(byts), "127.0.0.1") {
		t.Fatalf("backend config does not point at the proxy: %s", byts)
	}

	resp, err := nethttp.Get(backend.localBase + "/v1/terraform-backend?workspace_id=tfw123&job_id=job123")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != nethttp.StatusNoContent {
		t.Fatalf("proxy status %d", resp.StatusCode)
	}
	if sawToken != token || sawPath != "/v1/terraform-backend" {
		t.Fatalf("upstream saw token %q path %q", sawToken, sawPath)
	}

	other, err := nethttp.Get(backend.localBase + "/v1/runners/runner-id/jobs")
	if err != nil {
		t.Fatal(err)
	}
	other.Body.Close()
	if other.StatusCode != nethttp.StatusNotFound {
		t.Fatalf("non-backend path status %d", other.StatusCode)
	}
}
