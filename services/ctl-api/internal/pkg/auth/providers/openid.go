package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const (
	OpenIDProviderName = "openid"
)

// OpenIDProvider implements the Provider interface for generic OpenID Connect providers.
type OpenIDProvider struct {
	BaseProvider
	issuerURL       string
	discoveryConfig *OpenIDDiscoveryConfig
}

// OpenIDDiscoveryConfig holds the discovered OIDC configuration from the well-known endpoint.
type OpenIDDiscoveryConfig struct {
	Issuer                string   `json:"issuer"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	TokenEndpoint         string   `json:"token_endpoint"`
	UserinfoEndpoint      string   `json:"userinfo_endpoint"`
	JwksURI               string   `json:"jwks_uri"`
	ScopesSupported       []string `json:"scopes_supported"`
	ClaimsSupported       []string `json:"claims_supported"`
}

// NewOpenIDProvider creates a new OpenID Connect provider instance.
func NewOpenIDProvider() *OpenIDProvider {
	return &OpenIDProvider{
		BaseProvider: BaseProvider{
			name: OpenIDProviderName,
		},
	}
}

// Configure initializes the OpenID provider with the given configuration.
// If IssuerURL is provided, it will attempt OIDC discovery to auto-configure endpoints.
func (p *OpenIDProvider) Configure(cfg *ProviderConfig) error {
	if cfg.Logger != nil {
		p.log = cfg.Logger
	} else {
		p.log = zap.NewNop()
	}

	p.issuerURL = cfg.IssuerURL

	// If we have an issuer URL, try OIDC discovery
	if p.issuerURL != "" {
		if err := p.discover(context.Background()); err != nil {
			p.log.Warn("OIDC discovery failed, falling back to manual configuration",
				zap.Error(err),
				zap.String("issuer", p.issuerURL))
		} else {
			// Use discovered endpoints if not explicitly configured
			if cfg.AuthURL == "" && p.discoveryConfig != nil {
				cfg.AuthURL = p.discoveryConfig.AuthorizationEndpoint
			}
			if cfg.TokenURL == "" && p.discoveryConfig != nil {
				cfg.TokenURL = p.discoveryConfig.TokenEndpoint
			}
			if cfg.UserInfoURL == "" && p.discoveryConfig != nil {
				cfg.UserInfoURL = p.discoveryConfig.UserinfoEndpoint
			}
		}
	}

	// Validate required configuration
	if cfg.ClientID == "" {
		return fmt.Errorf("openid: client_id is required")
	}
	if cfg.AuthURL == "" {
		return fmt.Errorf("openid: auth_url is required (or provide issuer_url for discovery)")
	}
	if cfg.TokenURL == "" {
		return fmt.Errorf("openid: token_url is required (or provide issuer_url for discovery)")
	}

	// Set default scopes if not provided
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{"openid", "email", "profile"}
	}

	p.SetupOAuth2Config(cfg)
	p.name = OpenIDProviderName

	p.log.Info("OpenID provider configured",
		zap.String("auth_url", cfg.AuthURL),
		zap.String("token_url", cfg.TokenURL),
		zap.String("userinfo_url", cfg.UserInfoURL),
		zap.Strings("scopes", cfg.Scopes))

	return nil
}

// discover fetches the OpenID Connect discovery document from the well-known endpoint.
func (p *OpenIDProvider) discover(ctx context.Context) error {
	wellKnownURL := strings.TrimSuffix(p.issuerURL, "/") + "/.well-known/openid-configuration"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wellKnownURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create discovery request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch discovery document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discovery request failed with status: %d", resp.StatusCode)
	}

	var config OpenIDDiscoveryConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return fmt.Errorf("failed to decode discovery document: %w", err)
	}

	p.discoveryConfig = &config
	p.log.Debug("OIDC discovery successful",
		zap.String("issuer", config.Issuer),
		zap.String("auth_endpoint", config.AuthorizationEndpoint),
		zap.String("token_endpoint", config.TokenEndpoint),
		zap.String("userinfo_endpoint", config.UserinfoEndpoint))

	return nil
}

// GetUserInfo exchanges the authorization code for tokens and retrieves user information.
func (p *OpenIDProvider) GetUserInfo(ctx context.Context, r *http.Request, opts ...oauth2.AuthCodeOption) (*UserInfo, *ProviderTokens, error) {
	code := r.URL.Query().Get("code")
	if code == "" {
		return nil, nil, fmt.Errorf("openid: authorization code not found in request")
	}

	// Exchange the code for tokens
	client, _, ptokens, err := p.ExchangeCode(ctx, code, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("openid: %w", err)
	}

	p.log.Debug("token exchange successful",
		zap.Int("access_token_len", len(ptokens.AccessToken)),
		zap.Int("id_token_len", len(ptokens.IDToken)))

	// Fetch user info from the userinfo endpoint
	data, err := p.FetchUserInfo(ctx, client)
	if err != nil {
		return nil, ptokens, fmt.Errorf("openid: %w", err)
	}

	p.log.Debug("userinfo response", zap.String("body", string(data)))

	// Parse the userinfo response
	user, err := p.ParseUserInfo(data)
	if err != nil {
		return nil, ptokens, fmt.Errorf("openid: %w", err)
	}

	return user, ptokens, nil
}

// GetDiscoveryConfig returns the discovered OIDC configuration, if available.
func (p *OpenIDProvider) GetDiscoveryConfig() *OpenIDDiscoveryConfig {
	return p.discoveryConfig
}

// AuthCodeURL returns the URL to redirect the user to for authentication.
func (p *OpenIDProvider) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return p.oauth2Cfg.AuthCodeURL(state, opts...)
}
