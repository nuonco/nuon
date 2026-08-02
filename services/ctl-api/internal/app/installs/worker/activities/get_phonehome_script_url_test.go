package activities

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

// Precedence lives in the activity so the two stack-generation paths cannot disagree
// about which script an install gets. It is also the only thing keeping an edit to
// nuonco/runner's main from shipping itself to the fleet, hence the assertion that the
// default is a pinned tag.
func TestGetPhoneHomeScriptRawResolvesURL(t *testing.T) {
	newServer := func(t *testing.T, body string) *httptest.Server {
		t.Helper()
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(body))
		}))
		t.Cleanup(srv.Close)

		return srv
	}

	t.Run("the app override wins over the environment override", func(t *testing.T) {
		appSrv := newServer(t, "app-script")
		envSrv := newServer(t, "env-script")

		a := &Activities{cfg: &internal.Config{PhoneHomeScriptURL: envSrv.URL}}

		got, err := a.GetPhoneHomeScriptRaw(context.Background(), &GetPhoneHomeScriptRequest{URL: appSrv.URL})
		require.NoError(t, err)
		assert.Equal(t, "app-script", string(got))
	})

	t.Run("the environment override is used when the app sets none", func(t *testing.T) {
		envSrv := newServer(t, "env-script")

		a := &Activities{cfg: &internal.Config{PhoneHomeScriptURL: envSrv.URL}}

		got, err := a.GetPhoneHomeScriptRaw(context.Background(), &GetPhoneHomeScriptRequest{})
		require.NoError(t, err)
		assert.Equal(t, "env-script", string(got))
	})

	// A 404 body would otherwise be embedded as the Lambda's source and fail at
	// CreateStack in the customer's account rather than here.
	t.Run("a non-2xx response is an error, not a rendered 404 page", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("404: Not Found"))
		}))
		t.Cleanup(srv.Close)

		a := &Activities{cfg: &internal.Config{}}

		_, err := a.GetPhoneHomeScriptRaw(context.Background(), &GetPhoneHomeScriptRequest{URL: srv.URL})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})
}

func TestDefaultAWSPhoneHomeScriptIsPinned(t *testing.T) {
	assert.NotContains(t, DefaultAWSPhoneHomeScript, "refs/heads/",
		"the default must not track a branch: any commit to the script would ship itself to every org")
	assert.Contains(t, DefaultAWSPhoneHomeScript, "refs/tags/")
}
