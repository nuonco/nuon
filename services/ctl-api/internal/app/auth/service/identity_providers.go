package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/auth/providers"
)

var errIdentityProviderNotFound = errors.New("identity provider not found")

// getIdentityProviders returns every provider a user can sign in with: the env-configured provider
// first, then the enabled global providers from the database. Order is the sign-in page's button
// order.
func (s *service) getIdentityProviders(ctx context.Context) ([]*app.IdentityProvider, error) {
	defaultProvider, err := s.getDefaultIdentityProvider()
	if err != nil {
		return nil, err
	}
	allProviders := []*app.IdentityProvider{defaultProvider}

	dbProviders, err := s.getIdentityProvidersFromDB(ctx)
	if err != nil {
		// Log but don't fail - default provider is sufficient
		s.l.Warn("failed to load identity providers from database", zap.Error(err))
	} else {
		allProviders = append(allProviders, dbProviders...)
	}

	return allProviders, nil
}

// getDefaultIdentityProvider builds an IdentityProvider from environment variables.
// This provider is required and the service should not start without valid config.
func (s *service) getDefaultIdentityProvider() (*app.IdentityProvider, error) {
	providerType := s.cfg.NuonAuthProviderType
	if providerType == "" {
		return nil, fmt.Errorf("nuon_auth_provider_type is required")
	}

	var pType app.ProviderType
	switch providerType {
	case string(app.ProviderTypeOIDC):
		pType = app.ProviderTypeOIDC
	case string(app.ProviderTypeGoogle):
		pType = app.ProviderTypeGoogle
	case string(app.ProviderTypeGitHub):
		pType = app.ProviderTypeGitHub
	default:
		return nil, fmt.Errorf("invalid nuon_auth_provider_type: %s (must be oidc, google, or github)", providerType)
	}

	ip := &app.IdentityProvider{
		ID:           app.EnvIdentityProviderID(pType),
		ProviderType: pType,
		Name:         s.cfg.NuonAuthProviderName,
		Enabled:      true,
	}

	base := providers.BaseConfig{
		ClientID:     s.cfg.NuonAuthClientID,
		ClientSecret: s.cfg.NuonAuthClientSecret,
		RedirectURL:  s.cfg.NuonAuthRedirectURL,
	}

	switch pType {
	case app.ProviderTypeOIDC:
		cfg := &providers.OpenIDConfig{
			BaseConfig: base,
			IssuerURL:  s.cfg.NuonAuthIssuerURL,
		}
		if err := cfg.Validate(); err != nil {
			return nil, fmt.Errorf("invalid openid provider config: %w", err)
		}
		if err := ip.SetOpenIDConfig(cfg); err != nil {
			return nil, fmt.Errorf("failed to set openid config: %w", err)
		}

	case app.ProviderTypeGoogle:
		cfg := &providers.GoogleConfig{BaseConfig: base}
		if err := cfg.Validate(); err != nil {
			return nil, fmt.Errorf("invalid google provider config: %w", err)
		}
		if err := ip.SetGoogleConfig(cfg); err != nil {
			return nil, fmt.Errorf("failed to set google config: %w", err)
		}

	case app.ProviderTypeGitHub:
		cfg := &providers.GitHubConfig{BaseConfig: base}
		if err := cfg.Validate(); err != nil {
			return nil, fmt.Errorf("invalid github provider config: %w", err)
		}
		if err := ip.SetGitHubConfig(cfg); err != nil {
			return nil, fmt.Errorf("failed to set github config: %w", err)
		}
	}

	// the Set*Config helpers reset ProviderType, and the name is not part of the config blob
	ip.ProviderType = pType
	ip.Name = s.cfg.NuonAuthProviderName

	return ip, nil
}

// getIdentityProvidersFromDB fetches all enabled global identity providers from the database.
// Global providers have no org_id (NULL) and are available to all users.
func (s *service) getIdentityProvidersFromDB(ctx context.Context) ([]*app.IdentityProvider, error) {
	var dbProviders []*app.IdentityProvider
	err := s.db.WithContext(ctx).
		Where(&app.IdentityProvider{Enabled: true}).
		Where("org_id IS NULL").
		Order("created_at asc").
		Find(&dbProviders).Error
	if err != nil {
		return nil, err
	}
	return dbProviders, nil
}

// getIdentityProvider resolves the `provider` query param, which is a provider ID. A bare provider
// type is still accepted so that links minted before providers became individually addressable keep
// working; it resolves to the first enabled provider of that type.
func (s *service) getIdentityProvider(ctx context.Context, ref string) (*app.IdentityProvider, error) {
	if ref == "" {
		return nil, errIdentityProviderNotFound
	}

	defaultProvider, defaultErr := s.getDefaultIdentityProvider()
	if defaultErr == nil && defaultProvider.ID == ref {
		return defaultProvider, nil
	}

	if !app.IsEnvIdentityProviderID(ref) {
		var provider app.IdentityProvider
		err := s.db.WithContext(ctx).
			Where(&app.IdentityProvider{ID: ref, Enabled: true}).
			Where("org_id IS NULL").
			First(&provider).Error
		if err == nil {
			return &provider, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to look up identity provider: %w", err)
		}
	}

	return s.getIdentityProviderByType(ctx, app.ProviderType(ref))
}

// getIdentityProviderByType returns the first enabled identity provider of the given type,
// preferring the env-configured provider.
func (s *service) getIdentityProviderByType(ctx context.Context, providerType app.ProviderType) (*app.IdentityProvider, error) {
	defaultProvider, err := s.getDefaultIdentityProvider()
	if err == nil && defaultProvider.ProviderType == providerType {
		return defaultProvider, nil
	}

	var provider app.IdentityProvider
	err = s.db.WithContext(ctx).
		Where(&app.IdentityProvider{ProviderType: providerType, Enabled: true}).
		Where("org_id IS NULL").
		Order("created_at asc").
		First(&provider).Error
	if err == nil {
		return &provider, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to look up identity provider: %w", err)
	}

	return nil, fmt.Errorf("%w: %s", errIdentityProviderNotFound, providerType)
}

// createProviderFromIdentityProvider creates a configured Provider from an IdentityProvider model.
func (s *service) createProviderFromIdentityProvider(ip *app.IdentityProvider) (providers.Provider, error) {
	switch ip.ProviderType {
	case app.ProviderTypeOIDC:
		cfg, err := ip.GetOpenIDConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get openid config: %w", err)
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
			return nil, fmt.Errorf("failed to configure openid provider: %w", err)
		}
		return provider, nil

	case app.ProviderTypeGoogle:
		cfg, err := ip.GetGoogleConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get google config: %w", err)
		}
		provider := providers.NewGoogleProvider()
		if err := provider.Configure(&providers.ProviderConfig{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       cfg.Scopes,
			Logger:       s.l,
		}); err != nil {
			return nil, fmt.Errorf("failed to configure google provider: %w", err)
		}
		return provider, nil

	case app.ProviderTypeGitHub:
		cfg, err := ip.GetGitHubConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get github config: %w", err)
		}
		provider := providers.NewGitHubProvider()
		if err := provider.Configure(&providers.ProviderConfig{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       cfg.Scopes,
			Logger:       s.l,
		}); err != nil {
			return nil, fmt.Errorf("failed to configure github provider: %w", err)
		}
		return provider, nil

	default:
		return nil, fmt.Errorf("unsupported provider type: %s", ip.ProviderType)
	}
}
