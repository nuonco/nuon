package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// clearAmbientCredentials removes every credential source Resolve consults, so a test
// asserting a specific precedence step is not quietly satisfied by the developer's
// shell or by CI already being inside GitHub Actions.
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

// stubExchanger records what Resolve handed it and returns a fixed result.
type stubExchanger struct {
	token  string
	err    error
	called bool
	gotOrg string
	gotJWT string
}

func (s *stubExchanger) ExchangeOIDCToken(_ context.Context, orgID, jwt string) (string, error) {
	s.called = true
	s.gotOrg = orgID
	s.gotJWT = jwt
	return s.token, s.err
}

func TestResolvePrecedence(t *testing.T) {
	t.Run("explicit token wins over the environment", func(t *testing.T) {
		clearAmbientCredentials(t)
		t.Setenv("NUON_API_TOKEN", "from-env")

		ex := &stubExchanger{}
		got, err := Resolve(t.Context(), Options{APIToken: "explicit"}, ex)
		require.NoError(t, err)
		assert.Equal(t, "explicit", got)
		assert.False(t, ex.called, "a static token must not trigger an exchange")
	})

	t.Run("environment is used when no explicit token is set", func(t *testing.T) {
		clearAmbientCredentials(t)
		t.Setenv("NUON_API_TOKEN", "from-env")

		got, err := Resolve(t.Context(), Options{}, &stubExchanger{})
		require.NoError(t, err)
		assert.Equal(t, "from-env", got)
	})

	// Surrounding whitespace comes from shell quoting accidents and CI secret
	// interpolation, and an untrimmed token fails as an opaque 401.
	t.Run("a whitespace-only token is treated as absent", func(t *testing.T) {
		clearAmbientCredentials(t)
		t.Setenv("NUON_API_TOKEN", "   ")

		_, err := Resolve(t.Context(), Options{APIToken: "  "}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no credentials")
	})

	// The failure a customer is most likely to hit, so the message has to name every
	// way out rather than just saying "unauthorized".
	t.Run("no credentials at all is an actionable error", func(t *testing.T) {
		clearAmbientCredentials(t)

		_, err := Resolve(t.Context(), Options{}, &stubExchanger{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), APITokenEnvVar)
		assert.Contains(t, err.Error(), "id-token: write")
	})
}

func TestResolveOIDCPath(t *testing.T) {
	// Checked before the ID token is fetched: the exchange cannot succeed without an
	// org, and reporting it here says why rather than surfacing a generic auth failure.
	t.Run("an ambient OIDC token without an org id explains itself", func(t *testing.T) {
		clearAmbientCredentials(t)
		t.Setenv("NUON_OIDC_TOKEN", "a-jwt")

		_, err := Resolve(t.Context(), Options{}, &stubExchanger{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no org id is set")
	})

	t.Run("the org id falls back to the environment", func(t *testing.T) {
		clearAmbientCredentials(t)
		t.Setenv("NUON_OIDC_TOKEN", "a-jwt")
		t.Setenv("NUON_ORG_ID", "org-from-env")

		ex := &stubExchanger{token: "exchanged"}
		got, err := Resolve(t.Context(), Options{}, ex)
		require.NoError(t, err)
		assert.Equal(t, "exchanged", got)
		assert.Equal(t, "org-from-env", ex.gotOrg)
		assert.Equal(t, "a-jwt", ex.gotJWT)
	})

	t.Run("an explicit org id wins over the environment", func(t *testing.T) {
		clearAmbientCredentials(t)
		t.Setenv("NUON_OIDC_TOKEN", "a-jwt")
		t.Setenv("NUON_ORG_ID", "org-from-env")

		ex := &stubExchanger{token: "exchanged"}
		_, err := Resolve(t.Context(), Options{OrgID: "org-explicit"}, ex)
		require.NoError(t, err)
		assert.Equal(t, "org-explicit", ex.gotOrg)
	})

	// Names the source so the caller can tell an Actions problem from a file problem.
	t.Run("an exchange failure reports where the token came from", func(t *testing.T) {
		clearAmbientCredentials(t)
		t.Setenv("NUON_OIDC_TOKEN", "a-jwt")
		t.Setenv("NUON_ORG_ID", "org-1")

		ex := &stubExchanger{err: errors.New("boom")}
		_, err := Resolve(t.Context(), Options{}, ex)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exchange OIDC token")
		assert.Contains(t, err.Error(), "boom")
	})

	// A client that only supports static tokens should say so rather than panic.
	t.Run("a nil exchanger with an ambient token is a clear error", func(t *testing.T) {
		clearAmbientCredentials(t)
		t.Setenv("NUON_OIDC_TOKEN", "a-jwt")

		_, err := Resolve(t.Context(), Options{OrgID: "org-1"}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot exchange")
	})
}
