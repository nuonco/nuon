package service

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/auth/providers"
)

func (s *service) verifyProviderReachable(ip *app.IdentityProvider) error {
	if ip.ProviderType != app.ProviderTypeOIDC {
		return nil
	}

	cfg, err := ip.GetOpenIDConfig()
	if err != nil {
		return fmt.Errorf("invalid openid config: %w", err)
	}

	provider := providers.NewOpenIDProvider()
	if err := provider.Configure(&providers.ProviderConfig{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       cfg.Scopes,
		IssuerURL:    cfg.IssuerURL,
		AuthURL:      cfg.AuthURL,
		TokenURL:     cfg.TokenURL,
		UserInfoURL:  cfg.UserInfoURL,
		Logger:       s.l,
	}); err != nil {
		return fmt.Errorf("could not reach the identity provider at %s: %w", cfg.IssuerURL, err)
	}

	return nil
}
