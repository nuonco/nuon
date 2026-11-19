package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
)

var (
	errTokenNotFound = errors.New("token not found")
	errTokenExpired  = errors.New("token expired")
)

// TokenInfo represents the validated token information.
type TokenInfo struct {
	TokenID   string
	AccountID string
	Email     string
	Username  string
}

type accessToken struct {
	AccountID     string
	Role          string
	TokenType     app.TokenType
	TTL           time.Duration
	SourceTokenID string
}

func (s *service) createAccessToken(tx *gorm.DB, spec accessToken) (*app.Token, error) {
	client := account.New(account.Params{DB: tx})
	req := account.SessionTokenRequest{AccountID: spec.AccountID, Issuer: s.domain, TTL: spec.TTL, SourceTokenID: spec.SourceTokenID}
	if spec.TokenType == app.TokenTypeOAuth {
		return client.IssueOAuthAccessToken(tx.Statement.Context, req, spec.Role)
	}
	return client.IssueSessionToken(tx.Statement.Context, req)
}

func lockGrantSource(tx *gorm.DB, accountID, tokenID string, historical bool) error {
	err := account.LockGrantSource(tx, accountID, tokenID, historical)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errTokenNotFound
	}
	return err
}

func (s *service) createToken(tx *gorm.DB, account *app.Account, sourceTokenID string) (string, error) {
	token, err := s.createAccessToken(tx, accessToken{
		AccountID:     account.ID,
		TokenType:     app.TokenTypeNuon,
		TTL:           time.Duration(s.cfg.NuonAuthTokenTTL) * time.Minute,
		SourceTokenID: sourceTokenID,
	})
	if err != nil {
		return "", err
	}

	return token.Token, nil
}

// validateToken looks up a token in the database and returns the associated account info.
func (s *service) validateToken(tokenValue string) (*TokenInfo, error) {
	if tokenValue == "" {
		return nil, errTokenNotFound
	}

	var token app.Token
	err := s.db.
		Where(&app.Token{Token: tokenValue, TokenType: app.TokenTypeNuon}).
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
		TokenID:   token.ID,
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

	return account.New(account.Params{DB: s.db}).EndSession(context.Background(), tokenValue)
}
