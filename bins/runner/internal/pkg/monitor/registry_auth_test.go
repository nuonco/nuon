package monitor

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/registry"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/runner/settings"
)

func TestWriteRegistryAuthConfig(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "docker")
	path := filepath.Join(dir, "config.json")

	withAuth := &settings.Settings{ContainerImageRegistryAuth: &settings.RegistryAuth{
		Registry: "europe-west4-docker.pkg.dev",
		Username: "oauth2accesstoken",
		Password: "secret",
	}}
	require.NoError(t, writeRegistryAuthConfig(dir, withAuth))

	contents, err := os.ReadFile(path)
	require.NoError(t, err)
	want := base64.StdEncoding.EncodeToString([]byte("oauth2accesstoken:secret"))
	require.JSONEq(t, `{"auths":{"europe-west4-docker.pkg.dev":{"auth":"`+want+`"}}}`, string(contents))

	info, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	// unchanged settings must not rewrite the file
	past := time.Now().Add(-time.Hour)
	require.NoError(t, os.Chtimes(path, past, past))
	require.NoError(t, writeRegistryAuthConfig(dir, withAuth))
	info, err = os.Stat(path)
	require.NoError(t, err)
	require.WithinDuration(t, past, info.ModTime(), time.Second)

	// dropping the auth clears the stale credential
	require.NoError(t, writeRegistryAuthConfig(dir, &settings.Settings{}))
	contents, err = os.ReadFile(path)
	require.NoError(t, err)
	require.JSONEq(t, `{"auths":{}}`, string(contents))
}

func TestResolveRunnerImageDigestUsesRegistryAuth(t *testing.T) {
	reg := registry.New()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "oauth2accesstoken" || pass != "secret" {
			w.Header().Set("WWW-Authenticate", `Basic realm="registry"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		reg.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	host := strings.TrimPrefix(srv.URL, "http://")

	ref, err := name.ParseReference(host+"/runner:main", name.WeakValidation)
	require.NoError(t, err)
	img, err := random.Image(64, 1)
	require.NoError(t, err)
	basic := &authn.Basic{Username: "oauth2accesstoken", Password: "secret"}
	require.NoError(t, remote.Write(ref, img, remote.WithAuth(basic)))
	want, err := img.Digest()
	require.NoError(t, err)

	_, _, err = resolveRunnerImageDigest(context.Background(), ref, &settings.Settings{})
	require.Error(t, err, "anonymous HEAD should be rejected")

	s := &settings.Settings{ContainerImageRegistryAuth: &settings.RegistryAuth{
		Registry: host,
		Username: "oauth2accesstoken",
		Password: "secret",
	}}
	got, opts, err := resolveRunnerImageDigest(context.Background(), ref, s)
	require.NoError(t, err)
	require.Equal(t, want.String(), got)
	require.NotEmpty(t, opts)
}

func TestRegistryAuthenticatorMatchesRegistry(t *testing.T) {
	s := &settings.Settings{ContainerImageRegistryAuth: &settings.RegistryAuth{Registry: "europe-west4-docker.pkg.dev"}}

	gar, err := name.ParseReference("europe-west4-docker.pkg.dev/acme/acme-runner/runner:1")
	require.NoError(t, err)
	_, ok := registryAuthenticator(s, gar)
	require.True(t, ok)

	ecr, err := name.ParseReference("public.ecr.aws/p7e3r5y0/runner:1")
	require.NoError(t, err)
	_, ok = registryAuthenticator(s, ecr)
	require.False(t, ok)
}

func TestWriteRegistryAuthConfigKeepsTokenAcrossReplicas(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "docker")
	path := filepath.Join(dir, "config.json")
	token := func(password string, expiresIn time.Duration) *settings.Settings {
		return &settings.Settings{ContainerImageRegistryAuth: &settings.RegistryAuth{
			Registry:  "europe-west4-docker.pkg.dev",
			Username:  "oauth2accesstoken",
			Password:  password,
			ExpiresAt: time.Now().Add(expiresIn),
		}}
	}
	written := func() string {
		contents, err := os.ReadFile(path)
		require.NoError(t, err)
		var cfg dockerConfig
		require.NoError(t, json.Unmarshal(contents, &cfg))
		decoded, err := base64.StdEncoding.DecodeString(cfg.Auths["europe-west4-docker.pkg.dev"].Auth)
		require.NoError(t, err)
		return strings.TrimPrefix(string(decoded), "oauth2accesstoken:")
	}

	require.NoError(t, writeRegistryAuthConfig(dir, token("replica-a", time.Hour)))
	require.Equal(t, "replica-a", written())

	// another replica's fresh token must not replace one with plenty of time left
	require.NoError(t, writeRegistryAuthConfig(dir, token("replica-b", 55*time.Minute)))
	require.Equal(t, "replica-a", written())

	writtenRegistryAuthMu.Lock()
	nearExpiry := writtenRegistryAuth[dir]
	nearExpiry.ExpiresAt = time.Now().Add(10 * time.Minute)
	writtenRegistryAuth[dir] = nearExpiry
	writtenRegistryAuthMu.Unlock()

	// near expiry, a token that expires sooner is not an improvement
	require.NoError(t, writeRegistryAuthConfig(dir, token("replica-c", 5*time.Minute)))
	require.Equal(t, "replica-a", written())

	// near expiry, a later-expiring token replaces it
	require.NoError(t, writeRegistryAuthConfig(dir, token("replica-d", time.Hour)))
	require.Equal(t, "replica-d", written())

	// a missing file is rewritten even when the remembered token is still fresh
	require.NoError(t, os.Remove(path))
	require.NoError(t, writeRegistryAuthConfig(dir, token("replica-e", time.Hour)))
	require.Equal(t, "replica-e", written())
}

func TestKeepWrittenRegistryAuthRegistryChange(t *testing.T) {
	written := settings.RegistryAuth{Registry: "europe-west4-docker.pkg.dev", Username: "oauth2accesstoken", Password: "a", ExpiresAt: time.Now().Add(time.Hour)}
	offered := written
	offered.Registry = "us-central1-docker.pkg.dev"
	require.False(t, keepWrittenRegistryAuth(written, offered))

	noExpiry := settings.RegistryAuth{Registry: written.Registry, Username: written.Username, Password: "b"}
	require.False(t, keepWrittenRegistryAuth(written, noExpiry), "without expiries fall back to comparing tokens")
}
