package service

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const oauthAuthorizationCodeExpiry = 10 * time.Minute

// OAuthAuthorize handles GET /oauth/authorize — the OAuth 2.0 authorization-code
// endpoint with PKCE (RFC 6749 §4.1, RFC 7636). It validates the client request,
// persists it, and hands the user off to the existing identity-provider login
// flow, resuming at /oauth/finish once the user has authenticated.
func (s *service) OAuthAuthorize(c *gin.Context) {
	clientID := c.Query("client_id")
	redirectURI := c.Query("redirect_uri")

	// A bad client_id/redirect_uri must NOT redirect back (RFC 6749 §4.1.2.1) —
	// show an error page instead.
	if clientID == "" || redirectURI == "" {
		s.respondError(c, http.StatusBadRequest, fmt.Errorf("client_id and redirect_uri are required"))
		return
	}

	var client app.OAuthClient
	err := s.db.WithContext(c.Request.Context()).Where(&app.OAuthClient{ID: clientID}).First(&client).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		s.respondError(c, http.StatusBadRequest, fmt.Errorf("unknown client_id"))
		return
	}
	if err != nil {
		s.l.Error("failed to look up oauth client", zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, fmt.Errorf("failed to look up client"))
		return
	}
	if !client.AllowsRedirectURI(redirectURI) {
		s.respondError(c, http.StatusBadRequest, fmt.Errorf("redirect_uri is not registered for this client"))
		return
	}

	// From here, errors are reported by redirecting back to the client.
	responseType := c.Query("response_type")
	if responseType != "code" {
		s.redirectOAuthError(c, redirectURI, c.Query("state"), "unsupported_response_type", "only response_type=code is supported")
		return
	}

	codeChallenge := c.Query("code_challenge")
	codeChallengeMethod := c.Query("code_challenge_method")
	if codeChallenge == "" {
		s.redirectOAuthError(c, redirectURI, c.Query("state"), "invalid_request", "code_challenge is required (PKCE)")
		return
	}
	if codeChallengeMethod != "S256" {
		s.redirectOAuthError(c, redirectURI, c.Query("state"), "invalid_request", "code_challenge_method must be S256")
		return
	}

	requestID, err := generateStateNonce()
	if err != nil {
		s.l.Error("failed to generate oauth request id", zap.Error(err))
		s.redirectOAuthError(c, redirectURI, c.Query("state"), "server_error", "failed to start authorization")
		return
	}

	authCode := app.OAuthAuthorizationCode{
		RequestID:           requestID,
		ClientID:            client.ID,
		RedirectURI:         redirectURI,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		Scope:               c.Query("scope"),
		ClientState:         c.Query("state"),
		Resource:            c.Query("resource"),
		ExpiresAt:           time.Now().Add(oauthAuthorizationCodeExpiry),
	}
	if err := s.db.WithContext(c.Request.Context()).Create(&authCode).Error; err != nil {
		s.l.Error("failed to persist oauth authorization request", zap.Error(err))
		s.redirectOAuthError(c, redirectURI, authCode.ClientState, "server_error", "failed to start authorization")
		return
	}

	finishURL := fmt.Sprintf("%s/oauth/finish?rid=%s", s.oauthIssuer(), url.QueryEscape(requestID))

	// If the user already has a valid session, skip straight to finishing.
	if tokenValue := s.findToken(c); tokenValue != "" {
		if _, err := s.validateToken(tokenValue); err == nil {
			s.redirect302(c, finishURL)
			return
		}
	}

	loginURL := fmt.Sprintf("/login?provider=%s&url=%s",
		url.QueryEscape(s.cfg.NuonAuthProviderType),
		url.QueryEscape(finishURL))
	s.redirect302(c, loginURL)
}

// OAuthFinish handles GET /oauth/finish — resumes the authorization-code flow
// after the user has authenticated with the identity provider. Rather than
// issuing the code immediately, it renders a consent screen where the user
// picks which role (read-only or admin) the token should carry. The choice is
// submitted to POST /oauth/consent.
func (s *service) OAuthFinish(c *gin.Context) {
	requestID := c.Query("rid")
	authCode, ok := s.loadPendingAuthCode(c, requestID)
	if !ok {
		return
	}

	tokenInfo, ok := s.requireAuthSession(c, authCode)
	if !ok {
		return
	}

	c.HTML(http.StatusOK, "auth/oauth_consent.tmpl", gin.H{
		"RequestID":     requestID,
		"Email":         tokenInfo.Email,
		"ClientID":      authCode.ClientID,
		"AdminScope":    string(app.RoleTypeOrgAdmin),
		"ReadOnlyScope": string(app.RoleTypeOrgReadOnly),
		// Pre-select the scope the client requested, defaulting to read-only.
		"DefaultAdmin": authCode.Scope == string(app.RoleTypeOrgAdmin),
	})
}

// OAuthConsent handles POST /oauth/consent — the user has chosen a role on the
// consent screen. It records the selected scope, issues the authorization code,
// and redirects back to the client.
func (s *service) OAuthConsent(c *gin.Context) {
	requestID := c.PostForm("rid")
	authCode, ok := s.loadPendingAuthCode(c, requestID)
	if !ok {
		return
	}

	tokenInfo, ok := s.requireAuthSession(c, authCode)
	if !ok {
		return
	}

	scope := c.PostForm("scope")
	if scope != string(app.RoleTypeOrgAdmin) {
		scope = string(app.RoleTypeOrgReadOnly)
	}

	code, err := generateStateNonce()
	if err != nil {
		s.l.Error("failed to generate authorization code", zap.Error(err))
		s.redirectOAuthError(c, authCode.RedirectURI, authCode.ClientState, "server_error", "failed to issue code")
		return
	}

	authCode.Code = code
	authCode.AccountID = tokenInfo.AccountID
	authCode.Scope = scope
	if err := s.db.WithContext(c.Request.Context()).
		Model(&authCode).
		Select("code", "account_id", "scope").
		Updates(&authCode).Error; err != nil {
		s.l.Error("failed to issue authorization code", zap.Error(err))
		s.redirectOAuthError(c, authCode.RedirectURI, authCode.ClientState, "server_error", "failed to issue code")
		return
	}

	s.l.Info("oauth authorization code issued",
		zap.String("client_id", authCode.ClientID),
		zap.String("account_id", authCode.AccountID),
		zap.String("scope", scope))

	redirect := s.buildRedirectWithParams(authCode.RedirectURI, map[string]string{
		"code":  code,
		"state": authCode.ClientState,
	})
	s.redirect302(c, redirect)
}

// loadPendingAuthCode fetches and validates an in-flight authorization request
// by its request ID. It writes the appropriate error response and returns
// ok=false when the request is missing, already completed, or expired.
func (s *service) loadPendingAuthCode(c *gin.Context, requestID string) (app.OAuthAuthorizationCode, bool) {
	var authCode app.OAuthAuthorizationCode
	if requestID == "" {
		s.respondError(c, http.StatusBadRequest, fmt.Errorf("rid is required"))
		return authCode, false
	}

	err := s.db.WithContext(c.Request.Context()).Where(&app.OAuthAuthorizationCode{RequestID: requestID}).First(&authCode).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		s.respondError(c, http.StatusBadRequest, fmt.Errorf("unknown authorization request"))
		return authCode, false
	}
	if err != nil {
		s.l.Error("failed to look up oauth authorization request", zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, fmt.Errorf("failed to resume authorization"))
		return authCode, false
	}

	if authCode.Consumed || authCode.Code != "" {
		s.respondError(c, http.StatusBadRequest, fmt.Errorf("authorization request already completed"))
		return authCode, false
	}
	if time.Now().After(authCode.ExpiresAt) {
		s.redirectOAuthError(c, authCode.RedirectURI, authCode.ClientState, "access_denied", "authorization request expired")
		return authCode, false
	}

	return authCode, true
}

// requireAuthSession ensures the browser carries a valid login session. It
// writes a 401 and returns ok=false otherwise.
func (s *service) requireAuthSession(c *gin.Context, authCode app.OAuthAuthorizationCode) (*TokenInfo, bool) {
	tokenValue := s.findToken(c)
	if tokenValue == "" {
		s.respondError(c, http.StatusUnauthorized, fmt.Errorf("not authenticated"))
		return nil, false
	}
	tokenInfo, err := s.validateToken(tokenValue)
	if err != nil {
		s.l.Warn("invalid token at oauth consent", zap.Error(err))
		s.respondError(c, http.StatusUnauthorized, fmt.Errorf("not authenticated"))
		return nil, false
	}
	return tokenInfo, true
}

// redirectOAuthError redirects back to the client with an OAuth error (RFC 6749 §4.1.2.1).
func (s *service) redirectOAuthError(c *gin.Context, redirectURI, state, code, desc string) {
	params := map[string]string{"error": code, "error_description": desc}
	if state != "" {
		params["state"] = state
	}
	s.redirect302(c, s.buildRedirectWithParams(redirectURI, params))
}

// buildRedirectWithParams appends query params to a redirect URI, preserving any
// existing query string.
func (s *service) buildRedirectWithParams(redirectURI string, params map[string]string) string {
	u, err := url.Parse(redirectURI)
	if err != nil {
		return redirectURI
	}
	q := u.Query()
	for k, v := range params {
		if v != "" {
			q.Set(k, v)
		}
	}
	u.RawQuery = q.Encode()
	return u.String()
}
