package helpers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v50/github"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func TestGetVCSConnectionClientMissingInstallation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "/app/installations/123/access_tokens", r.URL.Path)
		http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	client := github.NewClient(server.Client())
	baseURL, err := url.Parse(server.URL + "/")
	require.NoError(t, err)
	client.BaseURL = baseURL

	helpers := &Helpers{ghClient: client}
	_, err = helpers.GetVCSConnectionClient(context.Background(), &app.VCSConnection{GithubInstallID: "123"})

	var notFoundErr stderr.ErrNotFound
	require.ErrorAs(t, err, &notFoundErr)
	var githubErr *github.ErrorResponse
	require.ErrorAs(t, err, &githubErr)
	require.Equal(t, http.StatusNotFound, githubErr.Response.StatusCode)
	require.Equal(t, "This GitHub installation is no longer available. Reconnect GitHub to restore repository access.", notFoundErr.Description)
}
