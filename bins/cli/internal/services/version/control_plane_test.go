package version

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchControlPlane(t *testing.T) {
	t.Run("reads both fields the control plane publishes", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/version", r.URL.Path)
			_, _ = w.Write([]byte(`{"version":"0.19.1109","git_ref":"0.19.1109","recommended_cli_version":"0.19.1102"}`))
		}))
		defer srv.Close()

		cp := FetchControlPlane(context.Background(), srv.URL)
		require.NotNil(t, cp)
		assert.Equal(t, "0.19.1109", cp.Version)
		assert.Equal(t, "0.19.1102", cp.RecommendedCLI)
	})

	// Callers only ever inform off this, so an old or unreachable control plane has to
	// degrade to "no information" rather than an error.
	t.Run("older control plane without the field", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"version":"0.19.1000","git_ref":"0.19.1000"}`))
		}))
		defer srv.Close()

		cp := FetchControlPlane(context.Background(), srv.URL)
		require.NotNil(t, cp)
		assert.Equal(t, "0.19.1000", cp.Version)
		assert.Empty(t, cp.RecommendedCLI)
	})

	t.Run("unreachable control plane", func(t *testing.T) {
		assert.Nil(t, FetchControlPlane(context.Background(), "http://127.0.0.1:9"))
	})

	t.Run("garbage response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`<html>nope</html>`))
		}))
		defer srv.Close()

		assert.Nil(t, FetchControlPlane(context.Background(), srv.URL))
	})
}
