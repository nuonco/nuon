package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

const (
	OpenIDProviderName = "openid"
)

type OpenIDProvider struct {
	BaseProvider
	issuerURL       string
	discoveryConfig *OpenIDDiscoveryConfig
}

type OpenIDDiscoveryConfig struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserinfoEndpoint                  string   `json:"userinfo_endpoint"`
	JwksURI                           string   `json:"jwks_uri"`
	ScopesSupported                   []string `json:"scopes_supported"`
	ClaimsSupported                   []string `json:"claims_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
}

func NewOpenIDProvider() *OpenIDProvider {
	return &OpenIDProvider{
		BaseProvider: BaseProvider{
			name: OpenIDProviderName,
		},
	}
}

func (p *OpenIDProvider) Configure(cfg *ProviderConfig) error {
	if cfg.Logger != nil {
		p.log = cfg.Logger
	} else {
		p.log = zap.NewNop()
	}

	p.issuerURL = cfg.IssuerURL

	if p.issuerURL != "" {
		if err := p.discover(context.Background()); err != nil {
			p.log.Warn("OIDC discovery failed, falling back to manual configuration",
				zap.Error(err),
				zap.String("issuer", p.issuerURL))
		} else {
			if cfg.AuthURL == "" && p.discoveryConfig != nil {
				cfg.AuthURL = p.discoveryConfig.AuthorizationEndpoint
			}
			if cfg.TokenURL == "" && p.discoveryConfig != nil {
				cfg.TokenURL = p.discoveryConfig.TokenEndpoint
			}
			if cfg.UserInfoURL == "" && p.discoveryConfig != nil {
				cfg.UserInfoURL = p.discoveryConfig.UserinfoEndpoint
			}
			if cfg.AuthStyle == 0 && p.discoveryConfig != nil {
				cfg.AuthStyle = resolveAuthStyle(p.discoveryConfig.TokenEndpointAuthMethodsSupported)
			}
		}
	}

	if cfg.ClientID == "" {
		return fmt.Errorf("openid: client_id is required")
	}
	if cfg.ClientSecret == "" {
		return fmt.Errorf("openid: client_secret is required")
	}
	if cfg.AuthURL == "" {
		return fmt.Errorf("openid: auth_url is required (or provide issuer_url for discovery)")
	}
	if cfg.TokenURL == "" {
		return fmt.Errorf("openid: token_url is required (or provide issuer_url for discovery)")
	}

	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{"openid", "email", "profile"}
	}

	p.SetupOAuth2Config(cfg)
	p.name = OpenIDProviderName

	p.log.Info("OpenID provider configured",
		zap.String("auth_url", cfg.AuthURL),
		zap.String("token_url", cfg.TokenURL),
		zap.String("userinfo_url", cfg.UserInfoURL),
		zap.String("auth_style", authStyleName(cfg.AuthStyle)),
		zap.Strings("scopes", cfg.Scopes))

	return nil
}

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

func (p *OpenIDProvider) GetUserInfo(ctx context.Context, r *http.Request, opts ...oauth2.AuthCodeOption) (*UserInfo, *ProviderTokens, error) {
	code := r.URL.Query().Get("code")
	if code == "" {
		return nil, nil, fmt.Errorf("openid: authorization code not found in request")
	}

	client, _, ptokens, err := p.ExchangeCode(ctx, code, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("openid: %w", err)
	}

	p.log.Debug("token exchange successful",
		zap.Int("access_token_len", len(ptokens.AccessToken)),
		zap.Int("id_token_len", len(ptokens.IDToken)))

	data, err := p.FetchUserInfo(ctx, client)
	if err != nil {
		return nil, ptokens, fmt.Errorf("openid: %w", err)
	}

	p.log.Debug("userinfo response", zap.String("body", string(data)))

	user, err := p.ParseUserInfo(data)
	if err != nil {
		return nil, ptokens, fmt.Errorf("openid: %w", err)
	}

	return user, ptokens, nil
}

func (p *OpenIDProvider) GetDiscoveryConfig() *OpenIDDiscoveryConfig {
	return p.discoveryConfig
}

func (p *OpenIDProvider) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return p.oauth2Cfg.AuthCodeURL(state, opts...)
}

func resolveAuthStyle(methods []string) oauth2.AuthStyle {
	if len(methods) == 0 {
		return oauth2.AuthStyleAutoDetect
	}
	if slices.Contains(methods, "client_secret_post") {
		return oauth2.AuthStyleInParams
	}
	if slices.Contains(methods, "client_secret_basic") {
		return oauth2.AuthStyleInHeader
	}
	return oauth2.AuthStyleAutoDetect
}

func authStyleName(s oauth2.AuthStyle) string {
	switch s {
	case oauth2.AuthStyleInParams:
		return "client_secret_post"
	case oauth2.AuthStyleInHeader:
		return "client_secret_basic"
	default:
		return "auto_detect"
	}
}
