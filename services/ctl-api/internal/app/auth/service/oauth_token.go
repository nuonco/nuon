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

	s.redeemOAuthGrant(c, &app.OAuthAuthorizationCode{}, authCode.ID, oauthGrant{
		AccountID:     authCode.AccountID,
		ClientID:      authCode.ClientID,
		Scope:         authCode.Scope,
		SourceTokenID: authCode.SourceTokenID,
	}, "authorization code already used")
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

	s.redeemOAuthGrant(c, &app.OAuthRefreshToken{}, refresh.ID, oauthGrant{
		AccountID:     refresh.AccountID,
		ClientID:      refresh.ClientID,
		Scope:         refresh.Scope,
		SourceTokenID: refresh.SourceTokenID,
	}, "refresh token already used")
}

type oauthGrant struct {
	AccountID     string
	ClientID      string
	Scope         string
	SourceTokenID string
}

var errOAuthGrantConsumed = errors.New("oauth grant already consumed")

func (s *service) redeemOAuthGrant(c *gin.Context, model any, grantID string, grant oauthGrant, consumedDesc string) {
	accessTTL := time.Duration(s.cfg.OAuthAccessTokenTTL) * time.Minute
	refreshTTL := time.Duration(s.cfg.OAuthRefreshTokenTTL) * time.Minute
	sourceType, sourceID := tokenSource(grant.SourceTokenID)

	var access *app.Token
	var refreshValue string
	err := s.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		if err := lockToken(tx.Unscoped(), grant.SourceTokenID); err != nil {
			return err
		}

		res := tx.Model(model).
			Where("id = ? AND consumed = ?", grantID, false).
			Update("consumed", true)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errOAuthGrantConsumed
		}

		var err error
		access, err = s.createAccessToken(tx, accessToken{
			AccountID:  grant.AccountID,
			Role:       oauthScopeToRole(grant.Scope),
			TokenType:  app.TokenTypeOAuth,
			TTL:        accessTTL,
			SourceType: sourceType,
			SourceID:   sourceID,
		})
		if err != nil {
			return err
		}

		refreshValue, err = generateStateNonce()
		if err != nil {
			return err
		}
		return tx.Create(&app.OAuthRefreshToken{
			Token:         refreshValue,
			ClientID:      grant.ClientID,
			Scope:         grant.Scope,
			AccountID:     grant.AccountID,
			ExpiresAt:     time.Now().Add(refreshTTL),
			SourceTokenID: grant.SourceTokenID,
		}).Error
	})
	if errors.Is(err, errOAuthGrantConsumed) {
		oauthError(c, http.StatusBadRequest, "invalid_grant", consumedDesc)
		return
	}
	if errors.Is(err, errTokenNotFound) {
		oauthError(c, http.StatusBadRequest, "invalid_grant", "credential source no longer exists")
		return
	}
	if err != nil {
		s.l.Error("failed to issue oauth tokens", zap.Error(err))
		oauthError(c, http.StatusInternalServerError, "server_error", "failed to issue token")
		return
	}

	s.l.Info("oauth tokens issued", zap.String("account_id", grant.AccountID), zap.String("client_id", grant.ClientID))

	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	c.JSON(http.StatusOK, gin.H{
		"access_token":  access.Token,
		"token_type":    "Bearer",
		"expires_in":    int(accessTTL.Seconds()),
		"refresh_token": refreshValue,
		"scope":         grant.Scope,
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
