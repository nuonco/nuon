package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

var (
	errTokenNotFound = errors.New("token not found")
	errTokenExpired  = errors.New("token expired")
)

// TokenInfo represents the validated token information.
type TokenInfo struct {
	AccountID string
	Email     string
	Username  string
}

// createToken creates a new auth token for the account and stores it in the database.
func (s *service) createToken(account *app.Account) (string, error) {
	now := time.Now()
	tokenValue := domains.NewUserTokenID()

	token := app.Token{
		CreatedByID: account.ID,
		AccountID:   account.ID,
		Token:       tokenValue,
		TokenType:   app.TokenTypeNuon,
		ExpiresAt:   now.Add(time.Duration(s.cfg.NuonAuthTokenTTL) * time.Minute),
		IssuedAt:    now,
		Issuer:      s.domain,
	}

	if err := s.db.Create(&token).Error; err != nil {
		return "", fmt.Errorf("failed to create token: %w", err)
	}

	return tokenValue, nil
}

// validateToken looks up a token in the database and returns the associated account info.
func (s *service) validateToken(tokenValue string) (*TokenInfo, error) {
	if tokenValue == "" {
		return nil, errTokenNotFound
	}

	var token app.Token
	err := s.db.
		Where(&app.Token{Token: tokenValue}).
		First(&token).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errTokenNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to lookup token: %w", err)
	}

	// Check expiry
	if time.Now().After(token.ExpiresAt) {
		return nil, errTokenExpired
	}

	// Look up the account
	var account app.Account
	err = s.db.
		Where("id = ?", token.AccountID).
		First(&account).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errTokenNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to lookup account: %w", err)
	}

	return &TokenInfo{
		AccountID: account.ID,
		Email:     account.Email,
		Username:  account.Email, // Account doesn't have a separate username field
	}, nil
}

func (s *service) findToken(c *gin.Context) string {
	if cookie, err := s.getCookie(c); err == nil && cookie != "" {
		return cookie
	}
	return ""
}

// deleteToken soft deletes a token from the database.
func (s *service) deleteToken(tokenValue string) error {
	if tokenValue == "" {
		return nil
	}

	return s.db.
		Where(&app.Token{Token: tokenValue}).
		Delete(&app.Token{}).Error
}
