package terraform

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestTerraformClient(url string) *TerraformClient {
	return &TerraformClient{
		cache:      Cache{ttl: time.Minute, store: make(map[string]CacheEntry)},
		httpClient: &http.Client{Transport: rewriteTransport(url)},
	}
}

func TestGetLatestVersion_ReturnsPinned(t *testing.T) {
	// GetLatestVersion is temporarily hard-coded to avoid the GitHub API's
	// 403 rate-limiting; it must not make any HTTP call.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("GetLatestVersion should not make an HTTP call while pinned")
	}))
	defer server.Close()

	version, err := newTestTerraformClient(server.URL).GetLatestVersion()
	require.NoError(t, err)
	assert.Equal(t, PinnedLatestVersion, version)
}

// The tests below cover FetchVersion, which is dormant while GetLatestVersion
// is pinned but retained so the GitHub-API path is ready to be restored.

func TestFetchVersion_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(GitHubRelease{TagName: "v1.9.0"})
	}))
	defer server.Close()

	version, err := newTestTerraformClient(server.URL).FetchVersion()
	require.NoError(t, err)
	assert.Equal(t, "1.9.0", version)
}

func TestFetchVersion_StripsvPrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(GitHubRelease{TagName: "v2.0.0"})
	}))
	defer server.Close()

	version, err := newTestTerraformClient(server.URL).FetchVersion()
	require.NoError(t, err)
	assert.Equal(t, "2.0.0", version)
}

func TestFetchVersion_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := newTestTerraformClient(server.URL).FetchVersion()
	require.Error(t, err)
}

func TestFetchVersion_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	_, err := newTestTerraformClient(server.URL).FetchVersion()
	require.Error(t, err)
}

// rewriteTransport redirects all requests to the given base URL,
// allowing the client to hit the test server regardless of the original host.
type rewriteTransport string

func (base rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.URL.Scheme = "http"
	req.URL.Host = string(base)[len("http://"):]
	return http.DefaultTransport.RoundTrip(req)
}
