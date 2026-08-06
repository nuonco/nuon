package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/nuonco/nuon/pkg/runner/airgap"
	"github.com/nuonco/nuon/pkg/runner/airgap/day2"
	"github.com/nuonco/nuon/pkg/runner/airgap/day2state"
)

func writeStateObject(t *testing.T, dir, key string, value any) {
	t.Helper()
	raw, err := json.Marshal(value)
	require.NoError(t, err)
	path := filepath.Join(dir, filepath.FromSlash(key))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, raw, 0o600))
}

func testPortal(t *testing.T) (*portalServer, string) {
	t.Helper()
	dir := t.TempDir()
	now := time.Now().UTC()
	writeStateObject(t, dir, day2.CatalogKey, day2.Catalog{
		SchemaVersion: day2.SchemaVersion,
		DeploymentID:  "dep-1",
		BundleDigest:  "sha256:bundle",
		Refs: []day2.CatalogRef{{
			ID:        "restart-api",
			Kind:      day2.RefKindAction,
			Name:      "Restart API",
			Component: "api",
		}},
	})
	writeStateObject(t, dir, "health/latest.json", airgap.HealthSnapshot{
		ObservedAt: now,
		Components: []airgap.ComponentHealth{{ComponentID: "cmp-1", ComponentName: "api", ComponentType: "helm", Health: "healthy"}},
	})
	writeStateObject(t, dir, day2.RunStatusKey("run-1"), day2.RunStatus{RunID: "run-1", RefID: "restart-api", StartedAt: now})
	p, err := newPortalServer(day2state.NewLocal(dir), "secret", "operator", map[string]bool{"127.0.0.1:1234": true}, zaptest.NewLogger(t))
	require.NoError(t, err)
	return p, dir
}

func TestPortalSecurityAndAPI(t *testing.T) {
	p, dir := testPortal(t)
	h := p.handler()

	request := httptest.NewRequest(http.MethodGet, "http://evil.example/api/catalog", nil)
	response := httptest.NewRecorder()
	h.ServeHTTP(response, request)
	require.Equal(t, http.StatusForbidden, response.Code)

	request = httptest.NewRequest(http.MethodPost, "http://127.0.0.1:1234/api/dispatch", strings.NewReader(`{"ref_id":"restart-api"}`))
	response = httptest.NewRecorder()
	h.ServeHTTP(response, request)
	require.Equal(t, http.StatusForbidden, response.Code)

	request = httptest.NewRequest(http.MethodPost, "http://127.0.0.1:1234/api/dispatch", strings.NewReader(`{"ref_id":"restart-api"}`))
	request.Header.Set("X-CSRF-Token", "secret")
	response = httptest.NewRecorder()
	h.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	var dispatched map[string]string
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &dispatched))
	require.NotEmpty(t, dispatched["dispatch_id"])
	raw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(day2.RequestKey(dispatched["dispatch_id"]))))
	require.NoError(t, err)
	var saved day2.Request
	require.NoError(t, json.Unmarshal(raw, &saved))
	require.Equal(t, day2.SourcePortal, saved.Source)
	require.Equal(t, "operator", saved.RequestedBy)
	require.Equal(t, "sha256:bundle", saved.BundleDigest)

	for _, path := range []string{"/api/catalog", "/api/health", "/api/runs", "/api/runs/run-1"} {
		request = httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234"+path, nil)
		response = httptest.NewRecorder()
		h.ServeHTTP(response, request)
		require.Equal(t, http.StatusOK, response.Code, path)
		require.True(t, json.Valid(response.Body.Bytes()), path)
		require.Empty(t, response.Header().Get("Access-Control-Allow-Origin"))
		require.Equal(t, "DENY", response.Header().Get("X-Frame-Options"))
		require.Contains(t, response.Header().Get("Content-Security-Policy"), "frame-ancestors 'none'")
	}
}

func TestPortalEmbedsSPAAndCSRFToken(t *testing.T) {
	p, _ := testPortal(t)
	request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:1234/runs", nil)
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), `content="secret"`)
	require.NotContains(t, response.Body.String(), "{{CSRF_TOKEN}}")
	require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
}

type failingPutState struct {
	day2state.State
}

func (f failingPutState) PutIfAbsent(context.Context, string, []byte) error {
	return day2state.ErrObjectExists
}

func TestPortalDispatchUsesConditionalWrite(t *testing.T) {
	p, _ := testPortal(t)
	p.store = failingPutState{State: p.store}
	request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:1234/api/dispatch", strings.NewReader(`{"ref_id":"restart-api"}`))
	request.Header.Set("X-CSRF-Token", "secret")
	response := httptest.NewRecorder()
	p.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.Contains(t, response.Body.String(), day2state.ErrObjectExists.Error())
}
