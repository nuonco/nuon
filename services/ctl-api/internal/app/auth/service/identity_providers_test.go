package service

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/auth/providers"
)

func testService(cfg *internal.Config) *service {
	return &service{cfg: cfg, l: zap.NewNop()}
}

func oidcConfig() *internal.Config {
	return &internal.Config{
		NuonAuthProviderType: string(app.ProviderTypeOIDC),
		NuonAuthClientID:     "client-id",
		NuonAuthClientSecret: "client-secret",
		NuonAuthIssuerURL:    "https://acme.auth0.com",
		NuonAuthRedirectURL:  "https://auth.example.com/auth",
	}
}

func TestEnvIdentityProviderID(t *testing.T) {
	require.Equal(t, "default-oidc", app.EnvIdentityProviderID(app.ProviderTypeOIDC))
	require.True(t, app.IsEnvIdentityProviderID("default-oidc"))
	require.True(t, app.IsEnvIdentityProviderID("default-github"))
	require.False(t, app.IsEnvIdentityProviderID("idp_abc123"))
}

func TestGetDefaultIdentityProvider(t *testing.T) {
	t.Run("oidc", func(t *testing.T) {
		ip, err := testService(oidcConfig()).getDefaultIdentityProvider()
		require.NoError(t, err)
		require.Equal(t, app.ProviderTypeOIDC, ip.ProviderType)
		require.Equal(t, "default-oidc", ip.ID)
		require.True(t, ip.Enabled)

		cfg, err := ip.GetOpenIDConfig()
		require.NoError(t, err)
		require.Equal(t, "https://acme.auth0.com", cfg.IssuerURL)
	})

	t.Run("name comes from config", func(t *testing.T) {
		cfg := oidcConfig()
		cfg.NuonAuthProviderName = "Acme SSO"

		ip, err := testService(cfg).getDefaultIdentityProvider()
		require.NoError(t, err)
		require.Equal(t, "Acme SSO", ip.Name)
		require.Equal(t, app.ProviderTypeOIDC, ip.ProviderType)
	})

	t.Run("unset type", func(t *testing.T) {
		_, err := testService(&internal.Config{}).getDefaultIdentityProvider()
		require.Error(t, err)
	})

	t.Run("unknown type", func(t *testing.T) {
		cfg := oidcConfig()
		cfg.NuonAuthProviderType = "saml"

		_, err := testService(cfg).getDefaultIdentityProvider()
		require.Error(t, err)
	})
}

func TestProviderDisplayName(t *testing.T) {
	testCases := []struct {
		name     string
		provider *app.IdentityProvider
		expected string
	}{
		{
			name:     "explicit name wins",
			provider: &app.IdentityProvider{ProviderType: app.ProviderTypeOIDC, Name: "Acme SSO"},
			expected: "Acme SSO",
		},
		{
			name:     "unnamed oidc falls back",
			provider: &app.IdentityProvider{ProviderType: app.ProviderTypeOIDC},
			expected: "Single Sign-On",
		},
		{
			name:     "google",
			provider: &app.IdentityProvider{ProviderType: app.ProviderTypeGoogle},
			expected: "Google",
		},
		{
			name:     "github",
			provider: &app.IdentityProvider{ProviderType: app.ProviderTypeGitHub},
			expected: "GitHub",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, providerDisplayName(tc.provider))
		})
	}
}

func TestProviderHint(t *testing.T) {
	oidcWithIssuer := func(issuer string) *app.IdentityProvider {
		ip := &app.IdentityProvider{}
		require.NoError(t, ip.SetOpenIDConfig(&providers.OpenIDConfig{IssuerURL: issuer}))
		return ip
	}

	require.Equal(t, "acme.auth0.com", providerHint(oidcWithIssuer("https://acme.auth0.com")))
	require.Equal(t, "login.microsoftonline.com",
		providerHint(oidcWithIssuer("https://login.microsoftonline.com/00000000-0000-0000-0000-000000000000/v2.0")))

	// the sign-in page is unauthenticated, so a broken issuer must degrade rather than error
	require.Empty(t, providerHint(oidcWithIssuer("://not a url")))
	require.Empty(t, providerHint(oidcWithIssuer("")))
	require.Empty(t, providerHint(&app.IdentityProvider{ProviderType: app.ProviderTypeGoogle}))
	require.Empty(t, providerHint(&app.IdentityProvider{ProviderType: app.ProviderTypeOIDC}))
}
