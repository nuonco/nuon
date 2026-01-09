package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

// BaseProvider provides common functionality for OAuth providers.
type BaseProvider struct {
	name        string
	oauth2Cfg   *oauth2.Config
	userInfoURL string
	claims      []string
	log         *zap.Logger
}

// Name returns the provider name.
func (b *BaseProvider) Name() string {
	return b.name
}

// SetupOAuth2Config initializes the OAuth2 configuration.
func (b *BaseProvider) SetupOAuth2Config(cfg *ProviderConfig) {
	b.oauth2Cfg = &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       cfg.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.AuthURL,
			TokenURL: cfg.TokenURL,
		},
	}
	b.userInfoURL = cfg.UserInfoURL
	b.claims = cfg.ClaimsToExtract
	b.log = cfg.Logger
	if b.log == nil {
		b.log = zap.NewNop()
	}
}

// GetOAuth2Config returns the OAuth2 configuration for generating auth URLs.
func (b *BaseProvider) GetOAuth2Config() *oauth2.Config {
	return b.oauth2Cfg
}

// ExchangeCode exchanges the authorization code for tokens and returns an HTTP client.
func (b *BaseProvider) ExchangeCode(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*http.Client, *oauth2.Token, *ProviderTokens, error) {
	token, err := b.oauth2Cfg.Exchange(ctx, code, opts...)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	ptokens := &ProviderTokens{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
	}

	if !token.Expiry.IsZero() {
		ptokens.Expiry = token.Expiry.Unix()
	}

	// Extract ID token if present (OIDC providers)
	if idToken := token.Extra("id_token"); idToken != nil {
		if idTokenStr, ok := idToken.(string); ok {
			ptokens.IDToken = idTokenStr
		}
	}

	client := b.oauth2Cfg.Client(ctx, token)
	return client, token, ptokens, nil
}

// FetchUserInfo fetches user information from the provider's userinfo endpoint.
func (b *BaseProvider) FetchUserInfo(ctx context.Context, client *http.Client) ([]byte, error) {
	if b.userInfoURL == "" {
		return nil, fmt.Errorf("userinfo URL not configured")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, b.userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo request failed with status: %d", resp.StatusCode)
	}

	var body []byte
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		body = append(body, buf[:n]...)
		if err != nil {
			break
		}
	}

	return body, nil
}

// MapClaims extracts configured claims from the raw response.
func (b *BaseProvider) MapClaims(data []byte, customClaims *CustomClaims) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	if customClaims.Claims == nil {
		customClaims.Claims = make(map[string]any)
	}

	// If no specific claims configured, keep all
	if len(b.claims) == 0 {
		customClaims.Claims = raw
		return nil
	}

	// Filter to only configured claims
	for _, claim := range b.claims {
		if val, ok := raw[claim]; ok {
			customClaims.Claims[claim] = val
		}
	}

	return nil
}

// ParseUserInfo unmarshals the userinfo response into a UserInfo struct.
func (b *BaseProvider) ParseUserInfo(data []byte) (*UserInfo, error) {
	var user UserInfo
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, fmt.Errorf("failed to parse userinfo: %w", err)
	}

	// Store raw claims for potential custom processing
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err == nil {
		user.RawClaims = raw
	}

	user.PrepareUserData()
	return &user, nil
}

// Logger returns the configured logger.
func (b *BaseProvider) Logger() *zap.Logger {
	return b.log
}
