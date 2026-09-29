package preflight

import (
	"context"
	"fmt"

	internal "github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/auth/providers"
)

const (
	providerTypeOIDC   = "oidc"
	providerTypeGoogle = "google"
	providerTypeGitHub = "github"
)

var nuonAuthCheck = Check{
	Name:        "nuon-auth",
	Description: "nuon auth identity provider",

	Skip: func(cfg *internal.Config) (string, bool) {
		if cfg.NuonAuthIssuerURL == "" && cfg.NuonAuthProviderType == "" {
			return "nuon auth not configured", true
		}

		return "", false
	},

	Fields: func(cfg *internal.Config) []Field {
		return []Field{
			{Name: "nuon_auth_provider_type", Value: cfg.NuonAuthProviderType, Required: true},
			{
				Name:     "nuon_auth_issuer_url",
				Value:    cfg.NuonAuthIssuerURL,
				Required: needsIssuerURL(cfg.NuonAuthProviderType),
			},
			{Name: "nuon_auth_client_id", Value: cfg.NuonAuthClientID, Required: true},
			{Name: "nuon_auth_client_secret", Value: cfg.NuonAuthClientSecret, Required: true, Secret: true},
			{Name: "nuon_auth_redirect_url", Value: cfg.NuonAuthRedirectURL, Required: true},
			{Name: "nuon_auth_provider_name", Value: cfg.NuonAuthProviderName},
			{Name: "nuon_auth_session_key", Value: cfg.NuonAuthSessionKey, Required: true, Secret: true},
		}
	},

	Probe: func(_ context.Context, cfg *internal.Config) (string, error) {
		providerCfg := &providers.ProviderConfig{
			ClientID:     cfg.NuonAuthClientID,
			ClientSecret: cfg.NuonAuthClientSecret,
			RedirectURL:  cfg.NuonAuthRedirectURL,
			IssuerURL:    cfg.NuonAuthIssuerURL,
			Logger:       nopLogger(),
		}

		switch cfg.NuonAuthProviderType {
		case providerTypeGoogle:
			if err := providers.NewGoogleProvider().Configure(providerCfg); err != nil {
				return "", fmt.Errorf("google provider config invalid: %w", err)
			}

			return "", warnf("google credentials present but unverified: fixed endpoints, nothing to discover %s",
				summary("client_id", cfg.NuonAuthClientID))

		case providerTypeGitHub:
			if err := providers.NewGitHubProvider().Configure(providerCfg); err != nil {
				return "", fmt.Errorf("github provider config invalid: %w", err)
			}

			return "", warnf("github credentials present but unverified: fixed endpoints, nothing to discover %s",
				summary("client_id", cfg.NuonAuthClientID))

		case providerTypeOIDC:
			provider := providers.NewOpenIDProvider()
			if err := provider.Configure(providerCfg); err != nil {
				return "", fmt.Errorf("OIDC discovery failed: %w", err)
			}

			discovery := provider.GetDiscoveryConfig()
			if discovery == nil {
				return "", fmt.Errorf("OIDC discovery returned no configuration")
			}

			return fmt.Sprintf("OIDC discovery OK %s", summary("issuer", discovery.Issuer)), nil

		default:
			return "", fmt.Errorf("invalid nuon_auth_provider_type: %q (must be %s, %s, or %s)",
				cfg.NuonAuthProviderType, providerTypeOIDC, providerTypeGoogle, providerTypeGitHub)
		}
	},
}

func needsIssuerURL(providerType string) bool {
	switch providerType {
	case providerTypeGoogle, providerTypeGitHub:
		return false
	default:
		return true
	}
}
