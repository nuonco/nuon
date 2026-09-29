package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	deviceCodeExpiry = 5 * time.Minute
)

var deviceCodePattern = regexp.MustCompile(`^[A-Z0-9]{4}-[A-Z0-9]{4}$`)

var (
	errDeviceCodeMissing  = errors.New("device code is required")
	errDeviceCodeInvalid  = errors.New("invalid device code format, expected XXXX-XXXX")
	errDeviceCodeExpired  = errors.New("device code has expired")
	errDeviceCodeConsumed = errors.New("device code already used")
	errNotAuthenticated   = errors.New("you must be logged in to approve CLI access")
)

func validateDeviceCode(code string) error {
	if !deviceCodePattern.MatchString(code) {
		return errDeviceCodeInvalid
	}
	return nil
}

func (s *service) buildDeviceCodeURL(code string) string {
	if s.cfg.RootDomain == "localhost" {
		return fmt.Sprintf("http://localhost:8084/device/code?code=%s", code)
	}
	return fmt.Sprintf("https://%s/device/code?code=%s", s.domain, code)
}

func (s *service) DeviceCodePage(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		s.respondError(c, http.StatusBadRequest, errDeviceCodeMissing)
		return
	}

	if err := validateDeviceCode(code); err != nil {
		s.respondError(c, http.StatusBadRequest, err)
		return
	}

	tokenValue := s.findToken(c)
	if tokenValue == "" {
		s.redirect302(c, s.deviceCodeLoginURL(c.Request.Context(), code))
		return
	}

	tokenInfo, err := s.validateToken(tokenValue)
	if err != nil {
		s.l.Warn("invalid token in device code flow", zap.Error(err))
		s.clearCookie(c)
		s.redirect302(c, s.deviceCodeLoginURL(c.Request.Context(), code))
		return
	}

	c.HTML(http.StatusOK, "auth/device_approve.tmpl", gin.H{
		"Code":     code,
		"Email":    tokenInfo.Email,
		"Username": tokenInfo.Username,
	})
}

func (s *service) DeviceCodeApprove(c *gin.Context) {
	code := c.PostForm("code")
	if code == "" {
		s.respondError(c, http.StatusBadRequest, errDeviceCodeMissing)
		return
	}

	if err := validateDeviceCode(code); err != nil {
		s.respondError(c, http.StatusBadRequest, err)
		return
	}

	tokenValue := s.findToken(c)
	if tokenValue == "" {
		s.respondError(c, http.StatusUnauthorized, errNotAuthenticated)
		return
	}

	tokenInfo, err := s.validateToken(tokenValue)
	if err != nil {
		s.l.Warn("invalid token during device code approval", zap.Error(err))
		s.respondError(c, http.StatusUnauthorized, errNotAuthenticated)
		return
	}

	var account app.Account
	if err := s.db.Where("id = ?", tokenInfo.AccountID).First(&account).Error; err != nil {
		s.l.Error("failed to find account for device code approval",
			zap.String("account_id", tokenInfo.AccountID),
			zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, fmt.Errorf("failed to process approval"))
		return
	}

	var existing app.DeviceCode
	err = s.db.Where("code = ?", code).First(&existing).Error
	if err == nil {
		if existing.Consumed {
			s.respondError(c, http.StatusBadRequest, errDeviceCodeConsumed)
			return
		}
		if time.Now().After(existing.ExpiresAt) {
			s.respondError(c, http.StatusBadRequest, errDeviceCodeExpired)
			return
		}
		c.HTML(http.StatusOK, "auth/device_success.tmpl", gin.H{
			"Email": tokenInfo.Email,
		})
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		s.l.Error("database error checking device code", zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, fmt.Errorf("failed to process approval"))
		return
	}

	deviceCode := &app.DeviceCode{
		Code:      code,
		AccountID: account.ID,
		ExpiresAt: time.Now().Add(deviceCodeExpiry),
		Consumed:  false,
	}

	if err := s.db.Create(deviceCode).Error; err != nil {
		s.l.Error("failed to create device code",
			zap.String("account_id", account.ID),
			zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, fmt.Errorf("failed to save approval"))
		return
	}

	s.l.Info("device code approved",
		zap.String("account_id", account.ID),
		zap.String("email", account.Email))

	c.HTML(http.StatusOK, "auth/device_success.tmpl", gin.H{
		"Email": tokenInfo.Email,
	})
}

func (s *service) DeviceCodeToken(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "missing_code",
			"error_description": "device code is required",
		})
		return
	}

	if err := validateDeviceCode(code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_code",
			"error_description": err.Error(),
		})
		return
	}

	var deviceCode app.DeviceCode
	err := s.db.Where("code = ?", code).First(&deviceCode).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusOK, gin.H{
			"error":             "authorization_pending",
			"error_description": "waiting for user approval",
		})
		return
	}
	if err != nil {
		s.l.Error("database error looking up device code", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "failed to check device code",
		})
		return
	}

	if time.Now().After(deviceCode.ExpiresAt) {
		c.JSON(http.StatusOK, gin.H{
			"error":             "expired_token",
			"error_description": "device code has expired",
		})
		return
	}

	if deviceCode.Consumed {
		c.JSON(http.StatusOK, gin.H{
			"error":             "access_denied",
			"error_description": "device code has already been used",
		})
		return
	}

	var account app.Account
	if err := s.db.Where("id = ?", deviceCode.AccountID).First(&account).Error; err != nil {
		s.l.Error("failed to find account for device code",
			zap.String("account_id", deviceCode.AccountID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "failed to process token",
		})
		return
	}

	tokenValue, err := s.createToken(&account)
	if err != nil {
		s.l.Error("failed to create token for device code", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "failed to create token",
		})
		return
	}

	if err := s.db.Model(&deviceCode).Update("consumed", true).Error; err != nil {
		s.l.Error("failed to mark device code as consumed", zap.Error(err))
	}

	s.l.Info("device code token issued",
		zap.String("account_id", account.ID),
		zap.String("email", account.Email))

	c.JSON(http.StatusOK, gin.H{
		"access_token": tokenValue,
		"token_type":   "Bearer",
		"email":        account.Email,
	})
}

func (s *service) deviceCodeLoginURL(ctx context.Context, code string) string {
	returnURL := url.QueryEscape(s.buildDeviceCodeURL(code))

	identityProviders, err := s.getIdentityProviders(ctx)
	if err != nil || len(identityProviders) != 1 {
		if err != nil {
			s.l.Warn("failed to list identity providers for device code login", zap.Error(err))
		}
		return fmt.Sprintf("/?url=%s", returnURL)
	}

	return fmt.Sprintf("/login?provider=%s&url=%s", identityProviders[0].ID, returnURL)
}
