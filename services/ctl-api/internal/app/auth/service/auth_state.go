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

func (s *service) AuthState(c *gin.Context) {
	pathState := c.Param("state")
	if pathState == "" {
		s.respondError(c, http.StatusBadRequest, errInvalidState)
		return
	}

	sessionData, err := s.getSession(c)
	if err != nil {
		s.l.Error("failed to get session", zap.Error(err))
		s.respondError(c, http.StatusBadRequest, errSessionNotFound)
		return
	}

	if sessionData.State != pathState {
		s.l.Error("state mismatch",
			zap.String("stored", sessionData.State),
			zap.String("received", pathState))
		s.respondError(c, http.StatusBadRequest, errStateMismatch)
		return
	}

	queryState := c.Query("state")
	if queryState != pathState {
		s.l.Error("query state mismatch",
			zap.String("path", pathState),
			zap.String("query", queryState))
		s.respondError(c, http.StatusBadRequest, errStateMismatch)
		return
	}

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

	if !s.isEmailDomainAllowed(userInfo.Email) {
		s.l.Warn("authentication denied: email domain not allowed",
			zap.String("email", userInfo.Email),
			zap.String("provider_id", providerID))
		s.respondError(c, http.StatusForbidden, fmt.Errorf("access denied: your email domain is not authorized to use this service"))
		return
	}

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

	if err := s.verifyUser(userInfo); err != nil {
		s.l.Warn("user not authorized", zap.Error(err))
		s.respondError(c, http.StatusForbidden, fmt.Errorf("user not authorized: %w", err))
		return
	}

	tokenValue, err := s.createToken(account)
	if err != nil {
		s.l.Error("failed to create token", zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, fmt.Errorf("failed to create token: %w", err))
		return
	}

	s.setCookie(c, tokenValue)

	s.clearSession(c)

	if sessionData.RequestedURL != "" {
		s.l.Debug("redirecting to requested URL", zap.String("url", sessionData.RequestedURL))
		s.redirect302(c, sessionData.RequestedURL)
		return
	}

	s.redirect302(c, "/success")
}

func (s *service) verifyUser(userInfo *providers.UserInfo) error {
	if userInfo.Email == "" && userInfo.Username == "" {
		return fmt.Errorf("user has no email or username")
	}
	return nil
}

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

func (s *service) allowAllUsers(ip *app.IdentityProvider) bool {
	if ip.AllowAllUsers != nil {
		return *ip.AllowAllUsers
	}
	return s.cfg.NuonAuthAllowAllUsers
}
