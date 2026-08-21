package oidctoken

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{tokenEnvVar, tokenFileEnvVar, audienceEnvVar, ghaRequestURLEnvVar, ghaRequestTokenEnvVar} {
		t.Setenv(key, "")
		os.Unsetenv(key)
	}
}

func TestAudiencePrecedence(t *testing.T) {
	clearEnv(t)
	require.Equal(t, "https://api.nuon.co", Audience("", "https://api.nuon.co"))

	t.Setenv(audienceEnvVar, "from-env")
	require.Equal(t, "from-env", Audience("", "https://api.nuon.co"))
	require.Equal(t, "from-flag", Audience("from-flag", "https://api.nuon.co"))
}

func TestDetectPrecedence(t *testing.T) {
	ctx := context.Background()

	t.Run("no source", func(t *testing.T) {
		clearEnv(t)
		_, _, ok, err := Detect(ctx, "")
		require.NoError(t, err)
		require.False(t, ok)
	})

	t.Run("env token wins", func(t *testing.T) {
		clearEnv(t)
		t.Setenv(tokenEnvVar, "env-token")
		t.Setenv(tokenFileEnvVar, "/does/not/exist")

		token, source, ok, err := Detect(ctx, "")
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, "env-token", token)
		require.Equal(t, tokenEnvVar, source)
	})

	t.Run("token file", func(t *testing.T) {
		clearEnv(t)
		path := filepath.Join(t.TempDir(), "token")
		require.NoError(t, os.WriteFile(path, []byte("file-token\n"), 0o600))
		t.Setenv(tokenFileEnvVar, path)

		token, source, ok, err := Detect(ctx, "")
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, "file-token", token)
		require.Equal(t, tokenFileEnvVar, source)
	})

	t.Run("missing token file errors", func(t *testing.T) {
		clearEnv(t)
		t.Setenv(tokenFileEnvVar, "/does/not/exist")

		_, _, ok, err := Detect(ctx, "")
		require.True(t, ok)
		require.Error(t, err)
	})
}

func TestFetchGitHubActionsToken(t *testing.T) {
	ctx := context.Background()

	newActionsServer := func(t *testing.T, wantAudience string) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "Bearer runtime-token", r.Header.Get("Authorization"))
			require.Equal(t, wantAudience, r.URL.Query().Get("audience"))
			json.NewEncoder(w).Encode(map[string]string{"value": "gha-jwt"})
		}))
	}

	t.Run("fetches token with audience", func(t *testing.T) {
		clearEnv(t)
		server := newActionsServer(t, "https://api.nuon.co")
		defer server.Close()
		t.Setenv(ghaRequestURLEnvVar, server.URL+"?api-version=2")
		t.Setenv(ghaRequestTokenEnvVar, "runtime-token")

		token, source, ok, err := Detect(ctx, "https://api.nuon.co")
		require.NoError(t, err)
		require.True(t, ok)
		require.Equal(t, "gha-jwt", token)
		require.Equal(t, "GitHub Actions", source)
	})

	t.Run("server error surfaces", func(t *testing.T) {
		clearEnv(t)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer server.Close()
		t.Setenv(ghaRequestURLEnvVar, server.URL)
		t.Setenv(ghaRequestTokenEnvVar, "runtime-token")

		_, _, ok, err := Detect(ctx, "")
		require.True(t, ok)
		require.Error(t, err)
	})
}
