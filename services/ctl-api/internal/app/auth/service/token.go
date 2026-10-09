package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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
	AccountID  string
	OrgID      string
	Name       string
	Role       string
	TokenType  app.TokenType
	TTL        time.Duration
	SourceType app.TokenSourceType
	SourceID   string
}

func (s *service) createAccessToken(tx *gorm.DB, spec accessToken) (*app.Token, error) {
	now := time.Now()
	token := app.Token{
		Token:       domains.NewUserTokenID(),
		TokenType:   spec.TokenType,
		AccountID:   spec.AccountID,
		OrgID:       spec.OrgID,
		Name:        spec.Name,
		CreatedByID: spec.AccountID,
		Role:        spec.Role,
		Issuer:      s.domain,
		IssuedAt:    now,
		ExpiresAt:   now.Add(spec.TTL),
		SourceType:  spec.SourceType,
		SourceID:    spec.SourceID,
	}
	if err := tx.Create(&token).Error; err != nil {
		return nil, fmt.Errorf("failed to create token: %w", err)
	}

	return &token, nil
}

func tokenSource(sourceTokenID string) (app.TokenSourceType, string) {
	if sourceTokenID == "" {
		return "", ""
	}
	return app.TokenSourceTypeToken, sourceTokenID
}

func lockGrantSource(tx *gorm.DB, accountID, tokenID string, historical bool) error {
	var account app.Account
	err := tx.Clauses(clause.Locking{Strength: "KEY SHARE"}).
		Select("id").
		Where(app.Account{ID: accountID}).
		First(&account).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errTokenNotFound
	}
	if err != nil {
		return err
	}

	if tokenID == "" {
		return nil
	}
	if historical {
		tx = tx.Unscoped()
	}

	var token app.Token
	err = tx.Clauses(clause.Locking{Strength: "SHARE"}).
		Select("id").
		Where(app.Token{ID: tokenID}).
		First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errTokenNotFound
	}
	return err
}

func (s *service) createToken(tx *gorm.DB, account *app.Account, sourceTokenID string) (string, error) {
	sourceType, sourceID := tokenSource(sourceTokenID)
	token, err := s.createAccessToken(tx, accessToken{
		AccountID:  account.ID,
		TokenType:  app.TokenTypeNuon,
		TTL:        time.Duration(s.cfg.NuonAuthTokenTTL) * time.Minute,
		SourceType: sourceType,
		SourceID:   sourceID,
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

	return s.db.
		Where(&app.Token{Token: tokenValue}).
		Delete(&app.Token{}).Error
}
