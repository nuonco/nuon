package stack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/sdks/stack/models"
)

// The whole point of this change: the identifier moves out of the URL and a bearer
// token moves into the header. Asserting both together so neither can regress alone.
func TestFetchConfigAuthenticatesAndKeysOnInstallID(t *testing.T) {
	clearAmbientCredentials(t)

	var gotPath, gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")

		_ = json.NewEncoder(w).Encode(models.ServiceStackConfigResponse{
			Config: &models.AppInstallerSDKConfig{
				InstallID:    "inst-1",
				PhoneHomeURL: "https://runner.example.com/v1/stacks/inst-1/phone-home",
			},
		})
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
	assert.Equal(t, "https://runner.example.com/v1/stacks/inst-1/phone-home", cfg.PhoneHomeURL)
	assert.Equal(t, "inst-1", cfg.InstallID)
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
	}))
	defer srv.Close()

	_, err := FetchConfig(t.Context(), Options{APIURL: srv.URL, InstallID: "inst-1", APIToken: "bad"})
	require.Error(t, err)
	assert.Equal(t, 1, calls, "a 401 must not be retried")
}

// PhoneHome reports to the host the config named, so an install can be directed at
// local, stage, or a BYOC control plane without the SDK composing the URL.
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
		srv.URL+"/v1/stacks/inst-1/phone-home",
		map[string]any{"request_type": "Create"},
	)
	require.NoError(t, err)

	assert.Equal(t, "/v1/stacks/inst-1/phone-home", gotPath)
	assert.Equal(t, "Bearer tok-abc", gotAuth)
	assert.Equal(t, "Create", gotPayload["request_type"])
}

func TestPhoneHomeRequiresAURL(t *testing.T) {
	clearAmbientCredentials(t)

	err := PhoneHome(t.Context(), Options{APIURL: "https://x", InstallID: "i", APIToken: "t"}, "", nil)
	require.ErrorContains(t, err, "phone_home_url is required")
}

// A stale capability URL carries a phone_home_id in the path. Reporting against it
// would silently target the wrong route, so the mismatch has to be an error.
func TestPhoneHomeRejectsAMismatchedPath(t *testing.T) {
	clearAmbientCredentials(t)

	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	err := PhoneHome(t.Context(),
		Options{APIURL: srv.URL, InstallID: "inst-1", APIToken: "tok-abc"},
		srv.URL+"/v1/installs/inst-1/phone-home/ph-1",
		map[string]any{"request_type": "Create"},
	)
	require.ErrorContains(t, err, "/v1/stacks/inst-1/phone-home")
	assert.Equal(t, 0, calls, "a mismatched path must not be reported against")
}

// The host is environment-specific and legitimately differs from Options.APIURL;
// only the path is contractual.
func TestPhoneHomeAcceptsADifferentHost(t *testing.T) {
	clearAmbientCredentials(t)

	var gotPath string
	reportSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusCreated)
	}))
	defer reportSrv.Close()

	err := PhoneHome(t.Context(),
		Options{APIURL: "https://runner.example.com", InstallID: "inst-1", APIToken: "tok-abc"},
		reportSrv.URL+"/v1/stacks/inst-1/phone-home",
		map[string]any{"request_type": "Create"},
	)
	require.NoError(t, err)
	assert.Equal(t, "/v1/stacks/inst-1/phone-home", gotPath)
}
