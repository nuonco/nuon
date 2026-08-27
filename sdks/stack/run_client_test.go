package stack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/sdks/stack/internal/core"
)

// The whole point of this change: the identifier moves out of the URL and a bearer
// token moves into the header. Asserting both together so neither can regress alone.
func TestFetchConfigAuthenticatesAndKeysOnInstallID(t *testing.T) {
	clearAmbientCredentials(t)

	var gotPath, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")

		_ = json.NewEncoder(w).Encode(configResponse{Config: &core.Config{
			InstallID:    "inst-1",
			PhoneHomeURL: "https://runner.example.com/v1/installs/inst-1/phone-home/ph-1",
		}})
	}))
	defer srv.Close()

	cfg, err := FetchConfig(t.Context(), Options{
		APIURL:    srv.URL,
		InstallID: "inst-1",
		APIToken:  "tok-abc",
	})
	require.NoError(t, err)

	assert.Equal(t, "/v1/stacks/inst-1/config", gotPath)
	assert.Equal(t, "Bearer tok-abc", gotAuth)

	// Serving the phone-home URL here is what lets the module drop phone_home_id.
	assert.Equal(t, "https://runner.example.com/v1/installs/inst-1/phone-home/ph-1", cfg.PhoneHomeURL)
}

func TestFetchConfigRequiresAPIURLAndInstallID(t *testing.T) {
	clearAmbientCredentials(t)

	_, err := FetchConfig(t.Context(), Options{InstallID: "inst-1", APIToken: "t"})
	require.ErrorContains(t, err, "api_url is required")

	_, err = FetchConfig(t.Context(), Options{APIURL: "https://x", APIToken: "t"})
	require.ErrorContains(t, err, "install_id is required")
}

// A rejected credential is rejected identically on every attempt, so retrying only
// delays the error. Counting requests is the only way to prove it did not retry.
func TestFetchConfigDoesNotRetryOnUnauthorized(t *testing.T) {
	clearAmbientCredentials(t)

	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer srv.Close()

	_, err := FetchConfig(t.Context(), Options{APIURL: srv.URL, InstallID: "inst-1", APIToken: "bad"})
	require.Error(t, err)
	assert.Equal(t, 1, calls, "a 401 must not be retried")
}

// PhoneHome posts to the URL the config handed back rather than composing one, so the
// caller never needs to know the phone-home ID.
func TestPhoneHomePostsToTheGivenURL(t *testing.T) {
	clearAmbientCredentials(t)

	var gotPath, gotAuth string
	var gotPayload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotPayload))
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	err := PhoneHome(t.Context(),
		Options{APIURL: srv.URL, InstallID: "inst-1", APIToken: "tok-abc"},
		srv.URL+"/v1/installs/inst-1/phone-home/ph-1",
		map[string]any{"request_type": "Create"},
	)
	require.NoError(t, err)

	assert.Equal(t, "/v1/installs/inst-1/phone-home/ph-1", gotPath)
	assert.Equal(t, "Bearer tok-abc", gotAuth)
	assert.Equal(t, "Create", gotPayload["request_type"])
}

func TestPhoneHomeRequiresAURL(t *testing.T) {
	clearAmbientCredentials(t)

	err := PhoneHome(t.Context(), Options{APIURL: "https://x", InstallID: "i", APIToken: "t"}, "", nil)
	require.ErrorContains(t, err, "phone_home_url is required")
}
