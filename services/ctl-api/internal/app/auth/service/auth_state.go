package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/auth/providers"
)

// AuthState handles the /auth/:state endpoint.
// This is where we:
// 1. Validate the state parameter matches what we stored in the session
// 2. Look up the identity provider from the session
// 3. Exchange the authorization code for tokens
// 4. Fetch user information from the provider
// 5. Relate the user to an account via account_identities
// 6. Create a token and set the auth cookie
// 7. Redirect to the originally requested URL
func (s *service) AuthState(c *gin.Context) {
	// Get the state from URL path
	pathState := c.Param("state")
	if pathState == "" {
		s.respondError(c, http.StatusBadRequest, errInvalidState)
		return
	}

	// Get the session
	sessionData, err := s.getSession(c)
	if err != nil {
		s.l.Error("failed to get session", zap.Error(err))
		s.respondError(c, http.StatusBadRequest, errSessionNotFound)
		return
	}

	// Validate the state matches what we stored
	if sessionData.State != pathState {
		s.l.Error("state mismatch",
			zap.String("stored", sessionData.State),
			zap.String("received", pathState))
		s.respondError(c, http.StatusBadRequest, errStateMismatch)
		return
	}

	// Also verify query state matches (belt and suspenders)
	queryState := c.Query("state")
	if queryState != pathState {
		s.l.Error("query state mismatch",
			zap.String("path", pathState),
			zap.String("query", queryState))
		s.respondError(c, http.StatusBadRequest, errStateMismatch)
		return
	}

	// Get the provider the flow was started with
	providerID := sessionData.ProviderID
	if providerID == "" {
		s.l.Error("no provider in session")
		s.respondError(c, http.StatusBadRequest, fmt.Errorf("no provider in session"))
		return
	}

	identityProvider, err := s.getIdentityProvider(c.Request.Context(), providerID)
	if err != nil {
		s.l.Error("failed to get identity provider",
			zap.String("provider_id", providerID),
			zap.Error(err))
		s.respondError(c, http.StatusBadRequest, fmt.Errorf("invalid provider"))
		return
	}

	provider, err := s.createProviderFromIdentityProvider(identityProvider)
	if err != nil {
		s.l.Error("failed to create provider",
			zap.String("provider_id", providerID),
			zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, fmt.Errorf("failed to initialize provider"))
		return
	}

	// Get user info from the provider (this exchanges the code for tokens)
	userInfo, _, err := provider.GetUserInfo(c.Request.Context(), c.Request)
	if err != nil {
		s.l.Error("failed to get user info from provider", zap.Error(err))
		s.respondError(c, http.StatusBadRequest, fmt.Errorf("failed to get user info: %w", err))
		return
	}

	s.l.Info("user authenticated via IdP",
		zap.String("email", userInfo.Email),
		zap.String("username", userInfo.Username),
		zap.String("subject", userInfo.Subject))

	// Check if user's email domain is allowed
	if !s.isEmailDomainAllowed(userInfo.Email) {
		s.l.Warn("authentication denied: email domain not allowed",
			zap.String("email", userInfo.Email),
			zap.String("provider_id", providerID))
		s.respondError(c, http.StatusForbidden, fmt.Errorf("access denied: your email domain is not authorized to use this service"))
		return
	}

	// Look up or create account by (identity_provider_id, sub)
	account, err := s.resolveAccount(c.Request.Context(), identityProvider, userInfo)
	if err != nil {
		if err == ErrAccountNotAuthorized {
			s.l.Warn("authentication denied: no account or pending invite",
				zap.String("provider_id", providerID),
				zap.String("sub", userInfo.Subject),
				zap.String("email", userInfo.Email))
			s.respondError(c, http.StatusForbidden, fmt.Errorf("access denied: you must have an existing account or a pending invitation to sign in"))
			return
		}
		if err == ErrEmailDomainNotAllowed {
			s.l.Warn("authentication denied: email domain not allowed",
				zap.String("provider_id", providerID),
				zap.String("sub", userInfo.Subject),
				zap.String("email", userInfo.Email))
			s.respondError(c, http.StatusForbidden, fmt.Errorf("access denied: your email domain is not authorized to use this service"))
			return
		}
		s.l.Error("failed to get or create account",
			zap.String("provider_id", providerID),
			zap.String("sub", userInfo.Subject),
			zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, fmt.Errorf("failed to process account: %w", err))
		return
	}

	s.l.Info("user account resolved",
		zap.String("account_id", account.ID),
		zap.String("email", account.Email))

	s.recordMarketingConsent(c.Request.Context(), sessionData, account)

	// Verify/authorize the user against allowed domains, whitelists, etc.
	if err := s.verifyUser(userInfo); err != nil {
		s.l.Warn("user not authorized", zap.Error(err))
		s.respondError(c, http.StatusForbidden, fmt.Errorf("user not authorized: %w", err))
		return
	}

	// Create auth token for the cookie
	tokenValue, err := s.createToken(account)
	if err != nil {
		s.l.Error("failed to create token", zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, fmt.Errorf("failed to create token: %w", err))
		return
	}

	// Set the auth cookie
	s.setCookie(c, tokenValue)

	// Clear the session now that auth is complete
	s.clearSession(c)

	// Redirect to the originally requested URL, or show success page
	if sessionData.RequestedURL != "" {
		s.l.Debug("redirecting to requested URL", zap.String("url", sessionData.RequestedURL))
		s.redirect302(c, sessionData.RequestedURL)
		return
	}

	// No requested URL - redirect to success page
	s.redirect302(c, "/success")
}

// recordMarketingConsent persists an opted-in consent checkbox onto the account. It is set-only
// (an unchecked box on a later login is not a revocation) and never fails the login — losing a
// consent bit is acceptable, losing a sign-in is not.
func (s *service) recordMarketingConsent(ctx context.Context, sessionData *SessionData, account *app.Account) {
	if !sessionData.MarketingConsent || account.MarketingConsent {
		return
	}

	res := s.db.WithContext(ctx).
		Model(&app.Account{}).
		Where(app.Account{ID: account.ID}).
		Update("marketing_consent", true)
	if res.Error != nil {
		s.l.Warn("failed to record marketing consent",
			zap.String("account_id", account.ID),
			zap.Error(res.Error))
		return
	}

	s.l.Info("recorded marketing consent", zap.String("account_id", account.ID))
}

// verifyUser checks if the user is authorized to access the system.
// TODO: Implement domain whitelists, team checks, etc.
func (s *service) verifyUser(userInfo *providers.UserInfo) error {
	// For now, allow all authenticated users
	// TODO: Add checks for:
	// - Allowed email domains
	// - User whitelists
	// - Team memberships (for GitHub)
	if userInfo.Email == "" && userInfo.Username == "" {
		return fmt.Errorf("user has no email or username")
	}
	return nil
}

// resolveAccount maps an authenticated IdP user onto a Nuon account, creating one if this
// provider admits new users.
func (s *service) resolveAccount(
	ctx context.Context,
	identityProvider *app.IdentityProvider,
	userInfo *providers.UserInfo,
) (*app.Account, error) {
	if s.allowAllUsers(identityProvider) {
		return s.getOrCreateAccountByIdentity(ctx, identityProvider, userInfo)
	}
	return s.getOrCreateAccountByIdentityStrict(ctx, identityProvider, userInfo)
}

// allowAllUsers reports whether this provider admits any user with an allowed email domain, or
// only users who already have an account or a pending invite. A provider may override the
// deployment-wide nuon_auth_allow_all_users so that, for example, a contractor IdP stays
// invite-only while the staff IdP allows self-signup.
func (s *service) allowAllUsers(ip *app.IdentityProvider) bool {
	if ip.AllowAllUsers != nil {
		return *ip.AllowAllUsers
	}
	return s.cfg.NuonAuthAllowAllUsers
}
