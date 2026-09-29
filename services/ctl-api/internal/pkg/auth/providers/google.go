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

type GoogleProvider struct {
	BaseProvider
}

type GoogleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
	HostDomain    string `json:"hd"`
}

func NewGoogleProvider() *GoogleProvider {
	return &GoogleProvider{
		BaseProvider: BaseProvider{
			name: GoogleProviderName,
		},
	}
}

func (p *GoogleProvider) Configure(cfg *ProviderConfig) error {
	if cfg.Logger != nil {
		p.log = cfg.Logger
	} else {
		p.log = zap.NewNop()
	}

	if cfg.ClientID == "" {
		return fmt.Errorf("google: client_id is required")
	}
	if cfg.ClientSecret == "" {
		return fmt.Errorf("google: client_secret is required")
	}

	if cfg.AuthURL == "" {
		cfg.AuthURL = google.Endpoint.AuthURL
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = google.Endpoint.TokenURL
	}
	if cfg.UserInfoURL == "" {
		cfg.UserInfoURL = GoogleUserInfoURL
	}

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

func (p *GoogleProvider) GetUserInfo(ctx context.Context, r *http.Request, opts ...oauth2.AuthCodeOption) (*UserInfo, *ProviderTokens, error) {
	code := r.URL.Query().Get("code")
	if code == "" {
		return nil, nil, fmt.Errorf("google: authorization code not found in request")
	}

	client, _, ptokens, err := p.ExchangeCode(ctx, code, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("google: %w", err)
	}

	p.log.Debug("token exchange successful",
		zap.Int("access_token_len", len(ptokens.AccessToken)),
		zap.Int("id_token_len", len(ptokens.IDToken)))

	data, err := p.FetchUserInfo(ctx, client)
	if err != nil {
		return nil, ptokens, fmt.Errorf("google: %w", err)
	}

	p.log.Debug("userinfo response", zap.String("body", string(data)))

	var googleUser GoogleUserInfo
	if err := json.Unmarshal(data, &googleUser); err != nil {
		return nil, ptokens, fmt.Errorf("google: failed to parse userinfo: %w", err)
	}

	user := &UserInfo{
		Subject:        googleUser.Sub,
		Email:          googleUser.Email,
		EmailVerified:  googleUser.EmailVerified,
		Name:           googleUser.Name,
		Username:       googleUser.Email,
		Picture:        googleUser.Picture,
		ProviderUserID: googleUser.Sub,
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err == nil {
		user.RawClaims = raw
	}

	return user, ptokens, nil
}

func (p *GoogleProvider) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return p.oauth2Cfg.AuthCodeURL(state, opts...)
}
