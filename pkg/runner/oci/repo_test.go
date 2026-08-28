package oci

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"

	"github.com/nuonco/nuon/pkg/plugins/configs"
)

func TestGetRepoDoesNotShareAuthCache(t *testing.T) {
	password := "first"
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, requestPassword, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("Www-Authenticate", `Basic realm="registry"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if username != "user" || requestPassword != password {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]any{
			"name": "repository",
			"tags": []string{},
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	getRepo := func(password string) *remote.Repository {
		repo, err := GetRepo(context.Background(), &configs.OCIRegistryRepository{
			RegistryType: configs.OCIRegistryTypePrivateOCI,
			Repository:   "repository",
			LoginServer:  server.URL,
			OCIAuth: &configs.OCIRegistryAuth{
				Username: "user",
				Password: password,
			},
		})
		require.NoError(t, err)

		remoteRepo := repo.(*remote.Repository)
		remoteRepo.Client.(*auth.Client).Client = server.Client()
		return remoteRepo
	}

	firstRepo := getRepo("first")
	require.NoError(t, firstRepo.Tags(context.Background(), "", func(tags []string) error { return nil }))

	password = "second"
	secondRepo := getRepo("second")
	require.NoError(t, secondRepo.Tags(context.Background(), "", func(tags []string) error { return nil }))
}

// TestGetRepoAnonymousDoesNotUseSharedGlobalClient guards the fix for
// anonymous/public pulls leaking credentials through oras-go's process-global
// auth.DefaultClient/auth.DefaultCache. GetRepo must always install an isolated
// per-repo auth client with its own cache — never leave repo.Client nil (which
// falls back to the shared global) and never reuse auth.DefaultCache.
func TestGetRepoAnonymousDoesNotUseSharedGlobalClient(t *testing.T) {
	repo, err := GetRepo(context.Background(), &configs.OCIRegistryRepository{
		RegistryType: configs.OCIRegistryTypePublicOCI,
		Repository:   "public.ecr.aws/p7e3r5y0/kitchen-sink-ui",
		OCIAuth:      &configs.OCIRegistryAuth{},
	})
	require.NoError(t, err)

	remoteRepo := repo.(*remote.Repository)
	require.NotNil(t, remoteRepo.Client, "anonymous repo must get an isolated client, not fall back to auth.DefaultClient")

	authClient, ok := remoteRepo.Client.(*auth.Client)
	require.True(t, ok)
	require.NotSame(t, auth.DefaultClient, authClient, "must not reuse the process-global default client")
	require.NotNil(t, authClient.Cache, "anonymous repo must have its own cache")
	require.NotSame(t, auth.DefaultCache, authClient.Cache, "must not reuse the process-global default cache")
	require.Nil(t, authClient.Credential, "anonymous repo must not attach static credentials")
}

func TestGetRepoUsesPlainHTTPForLocalhostRegistry(t *testing.T) {
	repo, err := GetRepo(context.Background(), &configs.OCIRegistryRepository{
		RegistryType: configs.OCIRegistryTypePublicOCI,
		Repository:   "localhost:5005/runner",
	})
	require.NoError(t, err)

	remoteRepo := repo.(*remote.Repository)
	require.True(t, remoteRepo.PlainHTTP)
}
