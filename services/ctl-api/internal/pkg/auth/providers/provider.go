package providers

import (
	"context"
	"net/http"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

type Provider interface {
	Name() string

	Configure(cfg *ProviderConfig) error

	GetUserInfo(ctx context.Context, r *http.Request, opts ...oauth2.AuthCodeOption) (*UserInfo, *ProviderTokens, error)
}

type ProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string

	AuthURL     string
	TokenURL    string
	UserInfoURL string

	AuthStyle oauth2.AuthStyle

	IssuerURL string

	ClaimsToExtract []string

	Logger *zap.Logger
}
