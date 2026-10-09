package service

import (
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oauthclients"
)

func (s *service) oauthTokenClientCredentials(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Pragma", "no-cache")
	invalidClient := func() {
		c.Header("WWW-Authenticate", `Basic realm="nuon"`)
		oauthError(c, http.StatusUnauthorized, "invalid_client", "client authentication failed")
	}

	rawID, rawSecret, ok := c.Request.BasicAuth()
	if !ok {
		invalidClient()
		return
	}
	clientID, err := url.QueryUnescape(rawID)
	if err != nil {
		invalidClient()
		return
	}
	plaintext, err := url.QueryUnescape(rawSecret)
	if err != nil {
		invalidClient()
		return
	}
	if c.Request.PostForm.Has("client_id") || c.Request.PostForm.Has("client_secret") {
		oauthError(c, http.StatusBadRequest, "invalid_request", "use only client_secret_basic for client authentication")
		return
	}

	ttl := time.Duration(s.cfg.OAuthAccessTokenTTL) * time.Minute
	var token *app.Token
	var client *app.OAuthClient
	var secret *app.OAuthClientSecret
	err = s.db.WithContext(c.Request.Context()).Transaction(func(tx *gorm.DB) error {
		var err error
		client, secret, err = oauthclients.Authenticate(tx, clientID, plaintext)
		if err != nil {
			return err
		}

		var acct app.Account
		if err := tx.Preload("Roles").Preload("Roles.Org").Preload("Roles.Policies").
			Where(app.Account{ID: client.AccountID.String}).First(&acct).Error; err != nil {
			return err
		}
		if !slices.Contains(acct.OrgIDs, client.OrgID.String) {
			return oauthclients.ErrInvalidClient
		}
		if strings.TrimSpace(c.PostForm("scope")) != "" || c.PostForm("resource") != "" {
			return errClientCredentialsScope
		}

		token, err = s.createAccessToken(tx, accessToken{
			AccountID:  acct.ID,
			OrgID:      client.OrgID.String,
			Name:       client.ClientName,
			TokenType:  app.TokenTypeOAuth,
			TTL:        ttl,
			SourceType: app.TokenSourceTypeOAuthClientSecret,
			SourceID:   secret.ID,
		})
		return err
	})
	if errors.Is(err, oauthclients.ErrInvalidClient) || errors.Is(err, gorm.ErrRecordNotFound) {
		invalidClient()
		return
	}
	if errors.Is(err, errClientCredentialsScope) {
		oauthError(c, http.StatusBadRequest, "invalid_scope", "scope and resource must be omitted for API access tokens")
		return
	}
	if err != nil {
		s.l.Error("failed to issue client credentials token", zap.String("client_id", clientID), zap.Error(err))
		oauthError(c, http.StatusInternalServerError, "server_error", "failed to issue token")
		return
	}

	now := time.Now()
	if err := s.db.WithContext(c.Request.Context()).Model(&app.OAuthClientSecret{}).
		Where(app.OAuthClientSecret{ID: secret.ID}).
		Where("last_used_at IS NULL OR last_used_at < ?", now.Add(-time.Minute)).
		Update("last_used_at", now).Error; err != nil {
		s.l.Warn("failed to update oauth client secret usage", zap.String("secret_id", secret.ID), zap.Error(err))
	}
	s.l.Info("client credentials token issued",
		zap.String("client_id", client.ID),
		zap.String("account_id", token.AccountID),
		zap.String("org_id", token.OrgID),
	)
	c.JSON(http.StatusOK, gin.H{
		"access_token": token.Token,
		"token_type":   "Bearer",
		"expires_in":   int(ttl.Seconds()),
	})
}

var errClientCredentialsScope = errors.New("unsupported client credentials scope or resource")
