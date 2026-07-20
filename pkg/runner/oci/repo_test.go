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
