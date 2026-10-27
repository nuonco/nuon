package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	deviceAuthorizationExpiry   = 10 * time.Minute
	deviceAuthorizationInterval = 5 * time.Second
	deviceCodeGrantType         = "urn:ietf:params:oauth:grant-type:device_code"

	// RFC 8628 §6.1: consonants only, so codes can't spell words and avoid look-alike characters.
	userCodeAlphabet = "BCDFGHJKLMNPQRSTVWXZ"
	userCodeLength   = 8

	userCodeAttempts = 5
)

var (
	errUserCodeMissing  = errors.New("enter the code shown in your terminal")
	errUserCodeNotFound = errors.New("that code is invalid or has expired")
	errNotAuthenticated = errors.New("you must be logged in to approve CLI access")
	errDeviceCodeUsed   = errors.New("device code already used")
	errLegacyDeviceFlow = errors.New("This version of the Nuon CLI is no longer supported. Update it with brew upgrade nuonco/tap/nuon, then run nuon login again.")
)

func generateUserCode() (string, error) {
	b := make([]byte, userCodeLength)
	max := big.NewInt(int64(len(userCodeAlphabet)))
	for i := range b {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		b[i] = userCodeAlphabet[n.Int64()]
	}
	return string(b[:4]) + "-" + string(b[4:]), nil
}

func generateDeviceCode() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// normalizeUserCode accepts what a user is likely to type (lowercase, spaces, missing dash) and
// returns the canonical XXXX-XXXX form, or "" if it can't be a user code.
func normalizeUserCode(input string) string {
	var b strings.Builder
	for _, r := range strings.ToUpper(input) {
		if r == '-' || r == ' ' {
			continue
		}
		if !strings.ContainsRune(userCodeAlphabet, r) {
			return ""
		}
		b.WriteRune(r)
	}
	code := b.String()
	if len(code) != userCodeLength {
		return ""
	}
	return code[:4] + "-" + code[4:]
}

func (s *service) buildDeviceURL() string {
	if s.cfg.RootDomain == "localhost" {
		return "http://localhost:8084/device"
	}
	return fmt.Sprintf("https://%s/device", s.domain)
}

func (s *service) buildDeviceURLComplete(userCode string) string {
	if userCode == "" {
		return s.buildDeviceURL()
	}
	return s.buildDeviceURL() + "?user_code=" + url.QueryEscape(userCode)
}

// DeviceAuthorize handles POST /device/authorize (RFC 8628 §3.1).
// The CLI starts the flow here and receives a device_code to poll with and a user_code to show the user.
func (s *service) DeviceAuthorize(c *gin.Context) {
	deviceCode, err := generateDeviceCode()
	if err != nil {
		s.l.Error("failed to generate device code", zap.Error(err))
		s.deviceTokenError(c, http.StatusInternalServerError, "server_error", "failed to start device authorization")
		return
	}

	var authz *app.DeviceAuthorization
	for range userCodeAttempts {
		userCode, err := generateUserCode()
		if err != nil {
			s.l.Error("failed to generate user code", zap.Error(err))
			break
		}

		candidate := &app.DeviceAuthorization{
			DeviceCode:         deviceCode,
			UserCode:           userCode,
			RequesterIP:        c.ClientIP(),
			RequesterUserAgent: c.Request.UserAgent(),
			Status:             app.DeviceAuthorizationStatusPending,
			ExpiresAt:          time.Now().Add(deviceAuthorizationExpiry),
		}
		err = s.db.WithContext(c.Request.Context()).Create(candidate).Error
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			continue
		}
		if err != nil {
			s.l.Error("failed to create device authorization", zap.Error(err))
			break
		}
		authz = candidate
		break
	}
	if authz == nil {
		s.deviceTokenError(c, http.StatusInternalServerError, "server_error", "failed to start device authorization")
		return
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{
		"device_code":               authz.DeviceCode,
		"user_code":                 authz.UserCode,
		"verification_uri":          s.buildDeviceURL(),
		"verification_uri_complete": s.buildDeviceURLComplete(authz.UserCode),
		"expires_in":                int(deviceAuthorizationExpiry.Seconds()),
		"interval":                  int(deviceAuthorizationInterval.Seconds()),
	})
}

// DevicePage handles GET /device
// Shows the form where a signed-in user enters the code from their terminal. A user_code in the
// query string (verification_uri_complete) only pre-fills the form; the user still has to submit
// it and then approve on the confirmation page.
func (s *service) DevicePage(c *gin.Context) {
	userCode := normalizeUserCode(c.Query("user_code"))

	tokenInfo, ok := s.deviceSession(c, userCode)
	if !ok {
		return
	}

	c.HTML(http.StatusOK, "auth/device_enter.tmpl", gin.H{
		"Email":    tokenInfo.Email,
		"UserCode": userCode,
	})
}

// DeviceLookup handles POST /device
// Finds the pending request for the entered code and asks the user to confirm it.
func (s *service) DeviceLookup(c *gin.Context) {
	tokenInfo, ok := s.deviceSession(c, "")
	if !ok {
		return
	}

	authz, err := s.findPendingDeviceAuthorization(c.Request.Context(), c.PostForm("user_code"))
	if err != nil {
		s.renderDeviceEnterError(c, tokenInfo, err)
		return
	}

	c.HTML(http.StatusOK, "auth/device_approve.tmpl", gin.H{
		"UserCode":    authz.UserCode,
		"Email":       tokenInfo.Email,
		"RequesterIP": authz.RequesterIP,
		"UserAgent":   authz.RequesterUserAgent,
		"RequestedAt": authz.CreatedAt.UTC().Format("2006-01-02 15:04:05 MST"),
	})
}

// DeviceApprove handles POST /device/approve
// Records the user's decision. Approval binds the request to the user's account; the CLI then
// redeems it at /device/token.
func (s *service) DeviceApprove(c *gin.Context) {
	tokenInfo, ok := s.deviceSession(c, "")
	if !ok {
		return
	}

	ctx := c.Request.Context()
	authz, err := s.findPendingDeviceAuthorization(ctx, c.PostForm("user_code"))
	if err != nil {
		s.renderDeviceEnterError(c, tokenInfo, err)
		return
	}

	if c.PostForm("action") != "approve" {
		if err := s.decideDeviceAuthorization(ctx, authz.ID, map[string]any{
			"status": app.DeviceAuthorizationStatusDenied,
		}); err != nil && !errors.Is(err, errUserCodeNotFound) {
			s.l.Error("failed to deny device authorization", zap.Error(err))
			s.respondError(c, http.StatusInternalServerError, fmt.Errorf("failed to save decision"))
			return
		}

		s.l.Info("device authorization denied", zap.String("account_id", tokenInfo.AccountID))
		c.HTML(http.StatusOK, "auth/device_denied.tmpl", gin.H{})
		return
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockGrantSource(tx, tokenInfo.AccountID, tokenInfo.TokenID, false); err != nil {
			return err
		}
		return s.decideDeviceAuthorizationTx(tx, authz.ID, map[string]any{
			"status":          app.DeviceAuthorizationStatusApproved,
			"account_id":      tokenInfo.AccountID,
			"source_token_id": tokenInfo.TokenID,
		})
	})
	if errors.Is(err, errUserCodeNotFound) {
		s.renderDeviceEnterError(c, tokenInfo, err)
		return
	}
	if errors.Is(err, errTokenNotFound) {
		s.respondError(c, http.StatusUnauthorized, errNotAuthenticated)
		return
	}
	if err != nil {
		s.l.Error("failed to approve device authorization",
			zap.String("account_id", tokenInfo.AccountID),
			zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, fmt.Errorf("failed to save approval"))
		return
	}

	s.l.Info("device authorization approved",
		zap.String("account_id", tokenInfo.AccountID),
		zap.String("email", tokenInfo.Email))

	c.HTML(http.StatusOK, "auth/device_success.tmpl", gin.H{
		"Email": tokenInfo.Email,
	})
}

// DeviceToken handles POST /device/token (RFC 8628 §3.4)
// The CLI polls with its device_code until the user approves or denies, or the code expires.
func (s *service) DeviceToken(c *gin.Context) {
	c.Header("Cache-Control", "no-store")

	if c.PostForm("grant_type") != deviceCodeGrantType {
		s.deviceTokenError(c, http.StatusBadRequest, "unsupported_grant_type", "grant_type must be "+deviceCodeGrantType)
		return
	}
	deviceCode := c.PostForm("device_code")
	if deviceCode == "" {
		s.deviceTokenError(c, http.StatusBadRequest, "invalid_request", "device_code is required")
		return
	}

	ctx := c.Request.Context()
	var authz app.DeviceAuthorization
	err := s.db.WithContext(ctx).
		Where(&app.DeviceAuthorization{DeviceCode: deviceCode}).
		First(&authz).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		s.deviceTokenError(c, http.StatusBadRequest, "invalid_grant", "unknown device code")
		return
	}
	if err != nil {
		s.l.Error("failed to look up device authorization", zap.Error(err))
		s.deviceTokenError(c, http.StatusInternalServerError, "server_error", "failed to check device code")
		return
	}

	switch {
	case authz.Status == app.DeviceAuthorizationStatusDenied:
		s.deviceTokenError(c, http.StatusBadRequest, "access_denied", "the request was denied")
		return
	case authz.Status == app.DeviceAuthorizationStatusConsumed:
		s.deviceTokenError(c, http.StatusBadRequest, "invalid_grant", "device code has already been used")
		return
	case time.Now().After(authz.ExpiresAt):
		s.deviceTokenError(c, http.StatusBadRequest, "expired_token", "device code has expired")
		return
	case authz.Status == app.DeviceAuthorizationStatusPending:
		s.deviceTokenError(c, http.StatusBadRequest, "authorization_pending", "waiting for user approval")
		return
	}

	var account app.Account
	if err := s.db.WithContext(ctx).Where(&app.Account{ID: authz.AccountID}).First(&account).Error; err != nil {
		s.l.Error("failed to find account for device authorization",
			zap.String("account_id", authz.AccountID),
			zap.Error(err))
		s.deviceTokenError(c, http.StatusBadRequest, "access_denied", "account no longer exists")
		return
	}

	var tokenValue string
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockGrantSource(tx, account.ID, authz.SourceTokenID, true); err != nil {
			return err
		}

		res := tx.Model(&app.DeviceAuthorization{}).
			Where("id = ? AND status = ?", authz.ID, app.DeviceAuthorizationStatusApproved).
			Update("status", app.DeviceAuthorizationStatusConsumed)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errDeviceCodeUsed
		}

		var err error
		tokenValue, err = s.createToken(tx, &account, authz.SourceTokenID)
		return err
	})
	if errors.Is(err, errDeviceCodeUsed) {
		s.deviceTokenError(c, http.StatusBadRequest, "invalid_grant", "device code has already been used")
		return
	}
	if errors.Is(err, errTokenNotFound) {
		s.deviceTokenError(c, http.StatusBadRequest, "access_denied", "credential source no longer exists")
		return
	}
	if err != nil {
		s.l.Error("failed to create token for device authorization", zap.Error(err))
		s.deviceTokenError(c, http.StatusInternalServerError, "server_error", "failed to create token")
		return
	}

	s.l.Info("device authorization token issued",
		zap.String("account_id", account.ID),
		zap.String("email", account.Email))

	c.JSON(http.StatusOK, gin.H{
		"access_token": tokenValue,
		"token_type":   "Bearer",
		"expires_in":   s.cfg.NuonAuthTokenTTL * 60,
		"email":        account.Email,
	})
}

// LegacyDevicePage handles GET /device/code from CLIs that still generate their own code.
func (s *service) LegacyDevicePage(c *gin.Context) {
	s.respondError(c, http.StatusGone, errLegacyDeviceFlow)
}

// LegacyDeviceToken handles GET /device/token from CLIs that still generate their own code.
func (s *service) LegacyDeviceToken(c *gin.Context) {
	s.deviceTokenError(c, http.StatusOK, "unsupported_client", errLegacyDeviceFlow.Error())
}

func (s *service) deviceTokenError(c *gin.Context, status int, code, description string) {
	c.JSON(status, gin.H{
		"error":             code,
		"error_description": description,
	})
}

// deviceSession returns the signed-in user, or redirects to login and returns false.
func (s *service) deviceSession(c *gin.Context, userCode string) (*TokenInfo, bool) {
	tokenValue := s.findToken(c)
	if tokenValue != "" {
		tokenInfo, err := s.validateToken(tokenValue)
		if err == nil {
			return tokenInfo, true
		}
		s.l.Warn("invalid token in device flow", zap.Error(err))
		s.clearCookie(c)
	}

	if c.Request.Method != http.MethodGet {
		s.respondError(c, http.StatusUnauthorized, errNotAuthenticated)
		return nil, false
	}
	s.redirect302(c, s.deviceLoginURL(c.Request.Context(), userCode))
	return nil, false
}

func (s *service) findPendingDeviceAuthorization(ctx context.Context, input string) (*app.DeviceAuthorization, error) {
	if strings.TrimSpace(input) == "" {
		return nil, errUserCodeMissing
	}
	userCode := normalizeUserCode(input)
	if userCode == "" {
		return nil, errUserCodeNotFound
	}

	var authz app.DeviceAuthorization
	err := s.db.WithContext(ctx).
		Where(&app.DeviceAuthorization{UserCode: userCode, Status: app.DeviceAuthorizationStatusPending}).
		First(&authz).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errUserCodeNotFound
	}
	if err != nil {
		return nil, err
	}
	if time.Now().After(authz.ExpiresAt) {
		return nil, errUserCodeNotFound
	}
	return &authz, nil
}

func (s *service) decideDeviceAuthorization(ctx context.Context, id string, updates map[string]any) error {
	return s.decideDeviceAuthorizationTx(s.db.WithContext(ctx), id, updates)
}

// decideDeviceAuthorizationTx moves a request out of pending only if it is still pending and
// unexpired, so two decisions can't race each other.
func (s *service) decideDeviceAuthorizationTx(tx *gorm.DB, id string, updates map[string]any) error {
	res := tx.Model(&app.DeviceAuthorization{}).
		Where("id = ? AND status = ? AND expires_at > ?", id, app.DeviceAuthorizationStatusPending, time.Now()).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errUserCodeNotFound
	}
	return nil
}

func (s *service) renderDeviceEnterError(c *gin.Context, tokenInfo *TokenInfo, err error) {
	if !errors.Is(err, errUserCodeMissing) && !errors.Is(err, errUserCodeNotFound) {
		s.l.Error("failed to look up device authorization", zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, fmt.Errorf("failed to look up code"))
		return
	}
	c.HTML(http.StatusBadRequest, "auth/device_enter.tmpl", gin.H{
		"Email": tokenInfo.Email,
		"Error": err.Error(),
	})
}

// deviceLoginURL sends the user straight to the only provider when there is one, and to the
// picker otherwise. Hardcoding the env provider here used to lock every CLI login onto it, leaving
// users whose account lives on another provider unable to complete `nuon auth login`.
func (s *service) deviceLoginURL(ctx context.Context, userCode string) string {
	returnURL := url.QueryEscape(s.buildDeviceURLComplete(userCode))

	identityProviders, err := s.getIdentityProviders(ctx)
	if err != nil || len(identityProviders) != 1 {
		if err != nil {
			s.l.Warn("failed to list identity providers for device login", zap.Error(err))
		}
		return fmt.Sprintf("/?url=%s", returnURL)
	}

	return fmt.Sprintf("/login?provider=%s&url=%s", identityProviders[0].ID, returnURL)
}
