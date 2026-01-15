package service

import (
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

// deviceCodePattern validates device codes in the format XXXX-XXXX where X is [A-Z0-9]
var deviceCodePattern = regexp.MustCompile(`^[A-Z0-9]{4}-[A-Z0-9]{4}$`)

var (
	errDeviceCodeMissing  = errors.New("device code is required")
	errDeviceCodeInvalid  = errors.New("invalid device code format, expected XXXX-XXXX")
	errDeviceCodeExpired  = errors.New("device code has expired")
	errDeviceCodeConsumed = errors.New("device code already used")
	errNotAuthenticated   = errors.New("you must be logged in to approve CLI access")
)

// validateDeviceCode checks if the code matches the expected format [A-Z0-9]{4}-[A-Z0-9]{4}
func validateDeviceCode(code string) error {
	if !deviceCodePattern.MatchString(code) {
		return errDeviceCodeInvalid
	}
	return nil
}

// buildDeviceCodeURL constructs the full device code URL with protocol.
// For localhost: http://localhost:8084/device/code?code=XXX
// Otherwise: https://auth.{RootDomain}/device/code?code=XXX
func (s *service) buildDeviceCodeURL(code string) string {
	if s.cfg.RootDomain == "localhost" {
		return fmt.Sprintf("http://localhost:8084/device/code?code=%s", code)
	}
	return fmt.Sprintf("https://%s/device/code?code=%s", s.domain, code)
}

// DeviceCodePage handles GET /device/code
// Displays the approval page for CLI authentication.
// Requires the user to be authenticated via X-Nuon-Auth cookie.
func (s *service) DeviceCodePage(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		s.respondError(c, http.StatusBadRequest, errDeviceCodeMissing)
		return
	}

	// Validate code format
	if err := validateDeviceCode(code); err != nil {
		s.respondError(c, http.StatusBadRequest, err)
		return
	}

	// Check if user is authenticated via cookie
	tokenValue := s.findToken(c)
	if tokenValue == "" {
		// User not logged in - redirect to login with return URL
		// Must use full URL with protocol for proper redirect after authentication
		returnURL := s.buildDeviceCodeURL(code)
		loginURL := fmt.Sprintf("/login?provider=%s&url=%s",
			s.cfg.NuonAuthProviderType,
			url.QueryEscape(returnURL))
		s.redirect302(c, loginURL)
		return
	}

	// Validate the token
	tokenInfo, err := s.validateToken(tokenValue)
	if err != nil {
		s.l.Warn("invalid token in device code flow", zap.Error(err))
		s.clearCookie(c)
		returnURL := s.buildDeviceCodeURL(code)
		loginURL := fmt.Sprintf("/login?provider=%s&url=%s",
			s.cfg.NuonAuthProviderType,
			url.QueryEscape(returnURL))
		s.redirect302(c, loginURL)
		return
	}

	// Render the approval page
	c.HTML(http.StatusOK, "auth/device_approve.tmpl", gin.H{
		"Code":     code,
		"Email":    tokenInfo.Email,
		"Username": tokenInfo.Username,
	})
}

// DeviceCodeApprove handles POST /device/code/approve
// Processes the user's approval and stores the code -> account mapping.
func (s *service) DeviceCodeApprove(c *gin.Context) {
	code := c.PostForm("code")
	if code == "" {
		s.respondError(c, http.StatusBadRequest, errDeviceCodeMissing)
		return
	}

	// Validate code format
	if err := validateDeviceCode(code); err != nil {
		s.respondError(c, http.StatusBadRequest, err)
		return
	}

	// Check if user is authenticated
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

	// Look up the account
	var account app.Account
	if err := s.db.Where("id = ?", tokenInfo.AccountID).First(&account).Error; err != nil {
		s.l.Error("failed to find account for device code approval",
			zap.String("account_id", tokenInfo.AccountID),
			zap.Error(err))
		s.respondError(c, http.StatusInternalServerError, fmt.Errorf("failed to process approval"))
		return
	}

	// Check if this code was already approved (prevent duplicates)
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
		// Already approved, show success
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

	// Create the approved device code record
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

// DeviceCodeToken handles GET /device/token
// CLI polls this endpoint to get a token once the code is approved.
func (s *service) DeviceCodeToken(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "missing_code",
			"error_description": "device code is required",
		})
		return
	}

	// Validate code format
	if err := validateDeviceCode(code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_code",
			"error_description": err.Error(),
		})
		return
	}

	// Look up the approved device code
	var deviceCode app.DeviceCode
	err := s.db.Where("code = ?", code).First(&deviceCode).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Code not approved yet - authorization_pending
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

	// Check if expired
	if time.Now().After(deviceCode.ExpiresAt) {
		c.JSON(http.StatusOK, gin.H{
			"error":             "expired_token",
			"error_description": "device code has expired",
		})
		return
	}

	// Check if already consumed
	if deviceCode.Consumed {
		c.JSON(http.StatusOK, gin.H{
			"error":             "access_denied",
			"error_description": "device code has already been used",
		})
		return
	}

	// Look up the account
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

	// Create a new API token for the CLI
	tokenValue, err := s.createToken(&account)
	if err != nil {
		s.l.Error("failed to create token for device code", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":             "server_error",
			"error_description": "failed to create token",
		})
		return
	}

	// Mark the device code as consumed
	if err := s.db.Model(&deviceCode).Update("consumed", true).Error; err != nil {
		s.l.Error("failed to mark device code as consumed", zap.Error(err))
		// Continue anyway - token was created successfully
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
