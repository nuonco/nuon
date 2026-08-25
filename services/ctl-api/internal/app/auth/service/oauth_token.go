package service

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// OAuthToken handles POST /oauth/token — the OAuth 2.0 token endpoint (RFC 6749).
// Supports the authorization_code grant (with PKCE) and the refresh_token grant.
func (s *service) OAuthToken(c *gin.Context) {
	switch c.PostForm("grant_type") {
	case "authorization_code":
		s.oauthTokenAuthorizationCode(c)
	case "refresh_token":
		s.oauthTokenRefresh(c)
	default:
		oauthError(c, http.StatusBadRequest, "unsupported_grant_type", "grant_type must be authorization_code or refresh_token")
	}
}

func (s *service) oauthTokenAuthorizationCode(c *gin.Context) {
	ctx := c.Request.Context()
	code := c.PostForm("code")
	clientID := c.PostForm("client_id")
	redirectURI := c.PostForm("redirect_uri")
	codeVerifier := c.PostForm("code_verifier")

	if code == "" || clientID == "" || codeVerifier == "" {
		oauthError(c, http.StatusBadRequest, "invalid_request", "code, client_id, and code_verifier are required")
		return
	}

	var authCode app.OAuthAuthorizationCode
	err := s.db.WithContext(ctx).Where(&app.OAuthAuthorizationCode{Code: code}).First(&authCode).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		oauthError(c, http.StatusBadRequest, "invalid_grant", "authorization code not found")
		return
	}
	if err != nil {
		s.l.Error("failed to look up authorization code", zap.Error(err))
		oauthError(c, http.StatusInternalServerError, "server_error", "failed to exchange code")
		return
	}

	if authCode.Consumed {
		oauthError(c, http.StatusBadRequest, "invalid_grant", "authorization code already used")
		return
	}
	if time.Now().After(authCode.ExpiresAt) {
		oauthError(c, http.StatusBadRequest, "invalid_grant", "authorization code expired")
		return
	}
	if authCode.ClientID != clientID {
		oauthError(c, http.StatusBadRequest, "invalid_grant", "client_id mismatch")
		return
	}
	if redirectURI != "" && authCode.RedirectURI != redirectURI {
		oauthError(c, http.StatusBadRequest, "invalid_grant", "redirect_uri mismatch")
		return
	}
	if !verifyPKCE(codeVerifier, authCode.CodeChallenge) {
		oauthError(c, http.StatusBadRequest, "invalid_grant", "PKCE verification failed")
		return
	}

	// Consume the code (single use).
	if err := s.db.WithContext(ctx).Model(&authCode).Update("consumed", true).Error; err != nil {
		s.l.Error("failed to consume authorization code", zap.Error(err))
		oauthError(c, http.StatusInternalServerError, "server_error", "failed to exchange code")
		return
	}

	s.issueOAuthTokens(c, authCode.AccountID, authCode.ClientID, authCode.Scope)
}

func (s *service) oauthTokenRefresh(c *gin.Context) {
	ctx := c.Request.Context()
	refreshValue := c.PostForm("refresh_token")
	clientID := c.PostForm("client_id")

	if refreshValue == "" || clientID == "" {
		oauthError(c, http.StatusBadRequest, "invalid_request", "refresh_token and client_id are required")
		return
	}

	var refresh app.OAuthRefreshToken
	err := s.db.WithContext(ctx).Where(&app.OAuthRefreshToken{Token: refreshValue}).First(&refresh).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		oauthError(c, http.StatusBadRequest, "invalid_grant", "refresh token not found")
		return
	}
	if err != nil {
		s.l.Error("failed to look up refresh token", zap.Error(err))
		oauthError(c, http.StatusInternalServerError, "server_error", "failed to refresh")
		return
	}

	if refresh.Consumed {
		oauthError(c, http.StatusBadRequest, "invalid_grant", "refresh token already used")
		return
	}
	if time.Now().After(refresh.ExpiresAt) {
		oauthError(c, http.StatusBadRequest, "invalid_grant", "refresh token expired")
		return
	}
	if refresh.ClientID != clientID {
		oauthError(c, http.StatusBadRequest, "invalid_grant", "client_id mismatch")
		return
	}

	// Rotate: consume the old refresh token before issuing a new pair.
	if err := s.db.WithContext(ctx).Model(&refresh).Update("consumed", true).Error; err != nil {
		s.l.Error("failed to rotate refresh token", zap.Error(err))
		oauthError(c, http.StatusInternalServerError, "server_error", "failed to refresh")
		return
	}

	s.issueOAuthTokens(c, refresh.AccountID, refresh.ClientID, refresh.Scope)
}

// issueOAuthTokens creates an access token (in the tokens table) and a refresh
// token, then writes the RFC 6749 token response.
func (s *service) issueOAuthTokens(c *gin.Context, accountID, clientID, scope string) {
	ctx := c.Request.Context()
	now := time.Now()

	accessTTL := time.Duration(s.cfg.OAuthAccessTokenTTL) * time.Minute
	refreshTTL := time.Duration(s.cfg.OAuthRefreshTokenTTL) * time.Minute

	accessToken := app.Token{
		Token:       domains.NewUserTokenID(),
		TokenType:   app.TokenTypeOAuth,
		Role:        oauthScopeToRole(scope),
		AccountID:   accountID,
		CreatedByID: accountID,
		Issuer:      s.domain,
		IssuedAt:    now,
		ExpiresAt:   now.Add(accessTTL),
	}
	if err := s.db.WithContext(ctx).Create(&accessToken).Error; err != nil {
		s.l.Error("failed to create oauth access token", zap.Error(err))
		oauthError(c, http.StatusInternalServerError, "server_error", "failed to issue token")
		return
	}

	refreshValue, err := generateStateNonce()
	if err != nil {
		s.l.Error("failed to generate refresh token", zap.Error(err))
		oauthError(c, http.StatusInternalServerError, "server_error", "failed to issue token")
		return
	}
	refresh := app.OAuthRefreshToken{
		Token:     refreshValue,
		ClientID:  clientID,
		Scope:     scope,
		AccountID: accountID,
		ExpiresAt: now.Add(refreshTTL),
	}
	if err := s.db.WithContext(ctx).Create(&refresh).Error; err != nil {
		s.l.Error("failed to create oauth refresh token", zap.Error(err))
		oauthError(c, http.StatusInternalServerError, "server_error", "failed to issue token")
		return
	}

	s.l.Info("oauth tokens issued", zap.String("account_id", accountID), zap.String("client_id", clientID))

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken.Token,
		"token_type":    "Bearer",
		"expires_in":    int(accessTTL.Seconds()),
		"refresh_token": refreshValue,
		"scope":         scope,
	})
}

// oauthScopeToRole maps a requested scope to the org role stored on the token.
// Only org_admin grants write access; everything else (including unknown/empty
// scopes) defaults to the least-privileged read-only role.
func oauthScopeToRole(scope string) string {
	if scope == string(app.RoleTypeOrgAdmin) {
		return string(app.RoleTypeOrgAdmin)
	}
	return string(app.RoleTypeOrgReadOnly)
}

// verifyPKCE checks an S256 PKCE code_verifier against the stored code_challenge
// (RFC 7636): base64url(sha256(verifier)) == challenge, compared in constant time.
func verifyPKCE(verifier, challenge string) bool {
	sum := sha256.Sum256([]byte(verifier))
	computed := base64.RawURLEncoding.EncodeToString(sum[:])
	return subtle.ConstantTimeCompare([]byte(computed), []byte(challenge)) == 1
}
