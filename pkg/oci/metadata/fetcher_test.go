package metadata

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

func TestDecodeInTotoStatement(t *testing.T) {
	statement := `{
		"_type": "https://in-toto.io/Statement/v0.1",
		"predicateType": "https://slsa.dev/provenance/v0.2",
		"subject": [{"name": "example/image"}],
		"predicate": {"materials": []}
	}`

	tests := []struct {
		name string
		data string
	}{
		{name: "direct", data: statement},
		{
			name: "DSSE envelope",
			data: fmt.Sprintf(`{"payloadType":"application/vnd.in-toto+json","payload":%q}`, base64.StdEncoding.EncodeToString([]byte(statement))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded, err := decodeInTotoStatement([]byte(tt.data))
			if err != nil {
				t.Fatalf("decodeInTotoStatement() error = %v", err)
			}
			if decoded.PredicateType != "https://slsa.dev/provenance/v0.2" {
				t.Fatalf("PredicateType = %q", decoded.PredicateType)
			}
			if len(decoded.Subject) != 1 || decoded.Subject[0].Name != "example/image" {
				t.Fatalf("Subject = %#v", decoded.Subject)
			}
		})
	}
}

// Guards the 12h-ECR-token replay: a shared cache pins the first token per host, and a
// nil client falls back to the same global cache for anonymous pulls.
func TestNewAuthClientDoesNotShareGlobalCache(t *testing.T) {
	authed := newAuthClient("registry.example.com", &RegistryAuth{Username: "AWS", Password: "first"})
	second := newAuthClient("registry.example.com", &RegistryAuth{Username: "AWS", Password: "second"})

	require.NotSame(t, auth.DefaultCache, authed.Cache, "must not reuse the process-global default cache")
	require.NotSame(t, authed.Cache, second.Cache, "each fetch must get its own cache")
	require.NotNil(t, authed.Credential)

	for name, regAuth := range map[string]*RegistryAuth{
		"nil auth":       nil,
		"empty username": {Username: ""},
	} {
		t.Run(name, func(t *testing.T) {
			anon := newAuthClient("public.ecr.aws", regAuth)
			require.NotNil(t, anon.Cache, "anonymous fetch must still get its own cache")
			require.NotSame(t, auth.DefaultCache, anon.Cache)
			require.Nil(t, anon.Credential, "anonymous fetch must not attach credentials")
		})
	}
}

// ECR advertises Basic auth, and oras-go caches Basic per host under a constant key,
// consulting the cache before Credential. A shared cache therefore replays the first
// token for a host even after the caller supplies a fresh one.
func TestNewAuthClientSendsRotatedCredential(t *testing.T) {
	want := "tokenA"
	var sent []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, pass, _ := r.BasicAuth()
		if pass == "" {
			w.Header().Set("WWW-Authenticate", `Basic realm="reg",service="ecr.amazonaws.com"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		sent = append(sent, pass)
		if pass != want {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"acme/app","tags":["v1"]}`))
	}))
	defer srv.Close()

	host := strings.TrimPrefix(srv.URL, "http://")
	probe := func(password string) error {
		repo, err := remote.NewRepository(host + "/acme/app")
		require.NoError(t, err)
		repo.PlainHTTP = true
		repo.Client = newAuthClient(host, &RegistryAuth{Username: "AWS", Password: password})
		return repo.Tags(context.Background(), "", func([]string) error { return nil })
	}

	require.NoError(t, probe("tokenA"))

	want = "tokenB" // as when ECR re-mints the 12h token
	require.NoError(t, probe("tokenB"), "must send the rotated credential, not a cached one")
	require.Equal(t, []string{"tokenA", "tokenB"}, sent)
}
