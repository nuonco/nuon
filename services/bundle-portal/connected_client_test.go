package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestConnectedPortalProxiesStandardAuthenticationServerSide(t *testing.T) {
	var authorization string
	var orgID string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		orgID = r.Header.Get("X-Nuon-Org-ID")
		require.Equal(t, "/v1/customer-managed/installs/ins-1/releases/rel-1", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"rel-1"}`))
	}))
	defer upstream.Close()
	client, err := newConnectedClient(upstream.URL, "org-1", "ins-1", "standard-token")
	require.NoError(t, err)
	portal, err := newPortalServer("csrf", map[string]bool{"portal.test": true}, zaptest.NewLogger(t))
	require.NoError(t, err)
	portal.connected = client

	request := httptest.NewRequest(http.MethodGet, "http://portal.test/api/connected/releases/rel-1", nil)
	response := httptest.NewRecorder()
	portal.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "Bearer standard-token", authorization)
	require.Equal(t, "org-1", orgID)
	require.NotContains(t, response.Body.String(), "standard-token")

	request = httptest.NewRequest(http.MethodGet, "http://portal.test/api/catalog", nil)
	response = httptest.NewRecorder()
	portal.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusNotFound, response.Code)
}

func TestConnectedPortalProxiesReleaseFileQuery(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/customer-managed/installs/ins-1/releases/rel-1/files/content", r.URL.Path)
		require.Equal(t, "components/lambda.toml", r.URL.Query().Get("path"))
		require.Equal(t, "Bearer standard-token", r.Header.Get("Authorization"))
		require.Equal(t, "org-1", r.Header.Get("X-Nuon-Org-ID"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"path":"components/lambda.toml","content":"name = \\"lambda\\""}`))
	}))
	defer upstream.Close()
	client, err := newConnectedClient(upstream.URL, "org-1", "ins-1", "standard-token")
	require.NoError(t, err)
	portal, err := newPortalServer("csrf", map[string]bool{"portal.test": true}, zaptest.NewLogger(t))
	require.NoError(t, err)
	portal.connected = client

	request := httptest.NewRequest(http.MethodGet, "http://portal.test/api/connected/releases/rel-1/files/content?path=components%2Flambda.toml", nil)
	response := httptest.NewRecorder()
	portal.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code)
}

func TestConnectedPortalProxiesWorkflowStepRetry(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/v1/customer-managed/installs/ins-1/workflows/wf-1/steps/step-1/retry", r.URL.Path)
		require.Equal(t, "Bearer standard-token", r.Header.Get("Authorization"))
		require.Equal(t, "org-1", r.Header.Get("X-Nuon-Org-ID"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"workflow_id":"wf-1","retryable":true}`))
	}))
	defer upstream.Close()
	client, err := newConnectedClient(upstream.URL, "org-1", "ins-1", "standard-token")
	require.NoError(t, err)
	portal, err := newPortalServer("csrf", map[string]bool{"portal.test": true}, zaptest.NewLogger(t))
	require.NoError(t, err)
	portal.connected = client

	request := httptest.NewRequest(http.MethodPost, "http://portal.test/api/connected/workflows/wf-1/steps/step-1/retry", strings.NewReader(`{}`))
	request.Header.Set("X-CSRF-Token", "csrf")
	response := httptest.NewRecorder()
	portal.handler().ServeHTTP(response, request)
	require.Equal(t, http.StatusCreated, response.Code)
	require.JSONEq(t, `{"workflow_id":"wf-1","retryable":true}`, response.Body.String())
}
