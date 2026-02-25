package providers

import (
	"context"
	"net/http"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

// Provider defines the interface that all OAuth/OIDC providers must implement.
type Provider interface {
	// Name returns the provider identifier (e.g., "google", "github", "openid").
	Name() string

	// Configure initializes the provider with the given configuration.
	Configure(cfg *ProviderConfig) error

	// GetUserInfo exchanges the authorization code for tokens and retrieves user information.
	GetUserInfo(ctx context.Context, r *http.Request, opts ...oauth2.AuthCodeOption) (*UserInfo, *ProviderTokens, error)
}

// ProviderConfig holds the configuration needed to set up an OAuth provider.
type ProviderConfig struct {
	// OAuth2 configuration
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string

	// Provider-specific URLs (some providers use discovery, others need explicit URLs)
	AuthURL     string
	TokenURL    string
	UserInfoURL string

	// AuthStyle specifies how the client credentials are sent to the token endpoint.
	// 0 = auto-detect (default), 1 = in params (POST body), 2 = in header (HTTP Basic).
	AuthStyle oauth2.AuthStyle

	// For OIDC providers
	IssuerURL string

	// Optional: claims to extract from the ID token or userinfo response
	ClaimsToExtract []string

	// Logger
	Logger *zap.Logger
}
