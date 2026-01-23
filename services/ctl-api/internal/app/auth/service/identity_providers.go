package service

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/auth/providers"
	"go.uber.org/zap"
)

const (
	// DefaultProviderID is the ID used for the env-configured default provider.
	DefaultProviderID = "default"
)

// getIdentityProviderIDPtr returns a pointer to the identity provider's ID,
// or nil if it's the default provider (which has no database record).
func getIdentityProviderIDPtr(ip *app.IdentityProvider) *string {
	if ip.ID == DefaultProviderID {
		return nil
	}
	return &ip.ID
}

// getIdentityProviders returns all configured identity providers from:
// 1. Environment variables (default provider - always present)
// 2. Database (additional configured providers)
//
// The default provider from env vars is always returned first and is guaranteed
// to exist if the service started successfully.
func (s *service) getIdentityProviders(ctx context.Context) ([]*app.IdentityProvider, error) {
	var allProviders []*app.IdentityProvider

	// 1. Get default provider from env vars
	defaultProvider, err := s.getDefaultIdentityProvider()
	if err != nil {
		return nil, err
	}
	allProviders = append(allProviders, defaultProvider)

	// 2. Get additional providers from database
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

	// Validate provider type
	var pType app.ProviderType
	switch providerType {
	case "oidc":
		pType = app.ProviderTypeOIDC
	case "google":
		pType = app.ProviderTypeGoogle
	case "github":
		pType = app.ProviderTypeGitHub
	default:
		return nil, fmt.Errorf("invalid nuon_auth_provider_type: %s (must be oidc, google, or github)", providerType)
	}

	// Build the identity provider based on type
	ip := &app.IdentityProvider{
		ID:           DefaultProviderID,
		ProviderType: pType,
		Enabled:      true,
	}

	// Set config based on provider type
	switch pType {
	case app.ProviderTypeOIDC:
		cfg := &providers.OpenIDConfig{
			BaseConfig: providers.BaseConfig{
				ClientID:     s.cfg.NuonAuthClientID,
				ClientSecret: s.cfg.NuonAuthClientSecret,
				RedirectURL:  s.cfg.NuonAuthRedirectURL,
			},
			IssuerURL: s.cfg.NuonAuthIssuerURL,
		}
		if err := cfg.Validate(); err != nil {
			return nil, fmt.Errorf("invalid openid provider config: %w", err)
		}
		if err := ip.SetOpenIDConfig(cfg); err != nil {
			return nil, fmt.Errorf("failed to set openid config: %w", err)
		}

	case app.ProviderTypeGoogle:
		cfg := &providers.GoogleConfig{
			BaseConfig: providers.BaseConfig{
				ClientID:     s.cfg.NuonAuthClientID,
				ClientSecret: s.cfg.NuonAuthClientSecret,
				RedirectURL:  s.cfg.NuonAuthRedirectURL,
			},
		}
		if err := cfg.Validate(); err != nil {
			return nil, fmt.Errorf("invalid google provider config: %w", err)
		}
		if err := ip.SetGoogleConfig(cfg); err != nil {
			return nil, fmt.Errorf("failed to set google config: %w", err)
		}

	case app.ProviderTypeGitHub:
		cfg := &providers.GitHubConfig{
			BaseConfig: providers.BaseConfig{
				ClientID:     s.cfg.NuonAuthClientID,
				ClientSecret: s.cfg.NuonAuthClientSecret,
				RedirectURL:  s.cfg.NuonAuthRedirectURL,
			},
		}
		if err := cfg.Validate(); err != nil {
			return nil, fmt.Errorf("invalid github provider config: %w", err)
		}
		if err := ip.SetGitHubConfig(cfg); err != nil {
			return nil, fmt.Errorf("failed to set github config: %w", err)
		}
	}

	return ip, nil
}

// getIdentityProvidersFromDB fetches all enabled global identity providers from the database.
// Global providers have no org_id (NULL) and are available to all users.
func (s *service) getIdentityProvidersFromDB(ctx context.Context) ([]*app.IdentityProvider, error) {
	var dbProviders []*app.IdentityProvider
	err := s.db.WithContext(ctx).
		Where("enabled = ? AND org_id IS NULL", true).
		Find(&dbProviders).Error
	if err != nil {
		return nil, err
	}
	return dbProviders, nil
}

// getIdentityProviderByType returns the first enabled identity provider of the given type.
// It checks the default provider first, then queries the database.
func (s *service) getIdentityProviderByType(ctx context.Context, providerType app.ProviderType) (*app.IdentityProvider, error) {
	// Check if default provider matches the requested type
	defaultProvider, err := s.getDefaultIdentityProvider()
	if err == nil && defaultProvider.ProviderType == providerType {
		return defaultProvider, nil
	}

	// Query database for provider by type
	var provider app.IdentityProvider
	err = s.db.WithContext(ctx).
		Where("provider_type = ? AND enabled = ?", providerType, true).
		First(&provider).Error
	if err == nil {
		return &provider, nil
	}

	return nil, fmt.Errorf("no identity provider found for type: %s", providerType)
}

// getProviderByType returns a configured Provider for the given ProviderType.
// This is the main helper for getting providers by type rather than by ID.
func (s *service) getProviderByType(ctx context.Context, providerType app.ProviderType) (providers.Provider, error) {
	ip, err := s.getIdentityProviderByType(ctx, providerType)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity provider: %w", err)
	}

	provider, err := s.createProviderFromIdentityProvider(ip)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return provider, nil
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
