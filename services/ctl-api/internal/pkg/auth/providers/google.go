package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	GoogleProviderName = "google"
	GoogleUserInfoURL  = "https://www.googleapis.com/oauth2/v3/userinfo"
)

// GoogleProvider implements the Provider interface for Google OAuth.
type GoogleProvider struct {
	BaseProvider
}

// GoogleUserInfo represents user information from Google's userinfo endpoint.
type GoogleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
	HostDomain    string `json:"hd"` // G Suite domain
}

// NewGoogleProvider creates a new Google OAuth provider instance.
func NewGoogleProvider() *GoogleProvider {
	return &GoogleProvider{
		BaseProvider: BaseProvider{
			name: GoogleProviderName,
		},
	}
}

// Configure initializes the Google provider with the given configuration.
func (p *GoogleProvider) Configure(cfg *ProviderConfig) error {
	if cfg.Logger != nil {
		p.log = cfg.Logger
	} else {
		p.log = zap.NewNop()
	}

	// Validate required configuration
	if cfg.ClientID == "" {
		return fmt.Errorf("google: client_id is required")
	}
	if cfg.ClientSecret == "" {
		return fmt.Errorf("google: client_secret is required")
	}

	// Set Google-specific defaults
	if cfg.AuthURL == "" {
		cfg.AuthURL = google.Endpoint.AuthURL
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = google.Endpoint.TokenURL
	}
	if cfg.UserInfoURL == "" {
		cfg.UserInfoURL = GoogleUserInfoURL
	}

	// Set default scopes if not provided
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{
			"openid",
			"email",
			"profile",
		}
	}

	p.SetupOAuth2Config(cfg)
	p.name = GoogleProviderName

	p.log.Info("Google provider configured",
		zap.String("userinfo_url", cfg.UserInfoURL),
		zap.Strings("scopes", cfg.Scopes))

	return nil
}

// GetUserInfo exchanges the authorization code for tokens and retrieves user information.
func (p *GoogleProvider) GetUserInfo(ctx context.Context, r *http.Request, opts ...oauth2.AuthCodeOption) (*UserInfo, *ProviderTokens, error) {
	code := r.URL.Query().Get("code")
	if code == "" {
		return nil, nil, fmt.Errorf("google: authorization code not found in request")
	}

	// Exchange the code for tokens
	client, _, ptokens, err := p.ExchangeCode(ctx, code, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("google: %w", err)
	}

	p.log.Debug("token exchange successful",
		zap.Int("access_token_len", len(ptokens.AccessToken)),
		zap.Int("id_token_len", len(ptokens.IDToken)))

	// Fetch user info from Google's userinfo endpoint
	data, err := p.FetchUserInfo(ctx, client)
	if err != nil {
		return nil, ptokens, fmt.Errorf("google: %w", err)
	}

	p.log.Debug("userinfo response", zap.String("body", string(data)))

	// Parse Google-specific response
	var googleUser GoogleUserInfo
	if err := json.Unmarshal(data, &googleUser); err != nil {
		return nil, ptokens, fmt.Errorf("google: failed to parse userinfo: %w", err)
	}

	// Convert to standard UserInfo
	user := &UserInfo{
		Subject:        googleUser.Sub,
		Email:          googleUser.Email,
		EmailVerified:  googleUser.EmailVerified,
		Name:           googleUser.Name,
		Username:       googleUser.Email, // Google uses email as username
		Picture:        googleUser.Picture,
		ProviderUserID: googleUser.Sub,
	}

	// Store raw claims
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err == nil {
		user.RawClaims = raw
	}

	return user, ptokens, nil
}

// AuthCodeURL returns the URL to redirect the user to for authentication.
func (p *GoogleProvider) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return p.oauth2Cfg.AuthCodeURL(state, opts...)
}
