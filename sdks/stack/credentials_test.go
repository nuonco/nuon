package stack

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// clearAmbientCredentials removes every credential source the resolver consults, so a
// test asserting a specific precedence step is not quietly satisfied by the
// developer's shell or by CI already being inside GitHub Actions.
func clearAmbientCredentials(t *testing.T) {
	t.Helper()

	for _, k := range []string{
		"NUON_API_TOKEN",
		"NUON_ORG_ID",
		"NUON_OIDC_TOKEN",
		"NUON_OIDC_TOKEN_FILE",
		"NUON_OIDC_AUDIENCE",
		"ACTIONS_ID_TOKEN_REQUEST_URL",
		"ACTIONS_ID_TOKEN_REQUEST_TOKEN",
	} {
		t.Setenv(k, "")
	}
}

func TestResolveTokenPrecedence(t *testing.T) {
	t.Run("explicit token wins over the environment", func(t *testing.T) {
		clearAmbientCredentials(t)
		t.Setenv("NUON_API_TOKEN", "from-env")

		got, err := resolveToken(t.Context(), Options{APIURL: "https://x", APIToken: "explicit"})
		require.NoError(t, err)
		assert.Equal(t, "explicit", got)
	})

	t.Run("environment is used when no explicit token is set", func(t *testing.T) {
		clearAmbientCredentials(t)
		t.Setenv("NUON_API_TOKEN", "from-env")

		got, err := resolveToken(t.Context(), Options{APIURL: "https://x"})
		require.NoError(t, err)
		assert.Equal(t, "from-env", got)
	})

	// The failure a customer is most likely to hit, so the message has to name every
	// way out rather than just saying "unauthorized".
	t.Run("no credentials at all is an actionable error", func(t *testing.T) {
		clearAmbientCredentials(t)

		_, err := resolveToken(t.Context(), Options{APIURL: "https://x"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "NUON_API_TOKEN")
		assert.Contains(t, err.Error(), "id-token: write")
	})

	// Checked before the ID token is fetched: the exchange cannot succeed without an
	// org, and reporting it here says why rather than surfacing a generic auth failure.
	t.Run("an ambient OIDC token without an org id explains itself", func(t *testing.T) {
		clearAmbientCredentials(t)
		t.Setenv("NUON_OIDC_TOKEN", "a-jwt")

		_, err := resolveToken(t.Context(), Options{APIURL: "https://x"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no org id is set")
	})
}

func TestResolveTokenExchangesOIDC(t *testing.T) {
	clearAmbientCredentials(t)
	t.Setenv("NUON_OIDC_TOKEN", "a-jwt")

	var gotBody exchangeRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/oidc/token", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		require.NoError(t, json.NewDecoder(r.Body).Decode(&gotBody))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(exchangeResponse{Authenticated: true, Token: "exchanged"})
	}))
	defer srv.Close()

	got, err := resolveToken(t.Context(), Options{APIURL: srv.URL, OrgID: "org-1"})
	require.NoError(t, err)
	assert.Equal(t, "exchanged", got)
	assert.Equal(t, "a-jwt", gotBody.Token, "the ID token must be forwarded verbatim")
	assert.Equal(t, "org-1", gotBody.OrgID)
}

// A 200 that did not authenticate is not a usable credential. The control plane
// returns a uniform failure for every auth-path rejection, so this has to be caught
// on the response body rather than the status code.
func TestResolveTokenRejectsUnauthenticatedExchange(t *testing.T) {
	clearAmbientCredentials(t)
	t.Setenv("NUON_OIDC_TOKEN", "a-jwt")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(exchangeResponse{Authenticated: false})
	}))
	defer srv.Close()

	_, err := resolveToken(t.Context(), Options{APIURL: srv.URL, OrgID: "org-1"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "trust policies")
}

// The audience has to match the trust policy's, and the dashboard creates policies
// with the public API URL while this SDK holds the runner API URL — so the value
// comes from NUON_OIDC_AUDIENCE, which the dashboard's directions print. Asserted
// through the real GitHub Actions path, the only caller that sends an audience.
func TestGitHubActionsAudienceFromEnv(t *testing.T) {
	clearAmbientCredentials(t)

	var gotAudience string
	actions := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAudience = r.URL.Query().Get("audience")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"value":"a-jwt"}`))
	}))
	defer actions.Close()

	exchange := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(exchangeResponse{Authenticated: true, Token: "exchanged"})
	}))
	defer exchange.Close()

	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_URL", actions.URL)
	t.Setenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN", "actions-token")
	t.Setenv("NUON_OIDC_AUDIENCE", "http://localhost:8081")

	got, err := resolveToken(t.Context(), Options{APIURL: exchange.URL, OrgID: "org-1"})
	require.NoError(t, err)
	assert.Equal(t, "exchanged", got)
	assert.Equal(t, "http://localhost:8081", gotAudience,
		"the env value must win over APIURL, which is the runner API and never the policy's audience")
}
