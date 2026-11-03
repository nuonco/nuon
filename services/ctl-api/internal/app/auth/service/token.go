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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

var (
	errTokenNotFound      = errors.New("token not found")
	errTokenExpired       = errors.New("token expired")
	errGrantAuthorization = errors.New("invalid grant authorization")
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
	Role       string
	TokenType  app.TokenType
	TTL        time.Duration
	SourceType app.TokenSourceType
	SourceID   string
	Limits     permissions.CredentialLimits
}

func (s *service) createAccessToken(tx *gorm.DB, spec accessToken) (*app.Token, error) {
	if spec.TTL <= 0 || spec.Limits.Profile != permissions.ProfileHuman || spec.Limits.API != permissions.APICtl {
		return nil, fmt.Errorf("invalid human access token specification")
	}
	now := time.Now()
	token := app.Token{
		Token:       domains.NewUserTokenID(),
		TokenType:   spec.TokenType,
		AccountID:   spec.AccountID,
		CreatedByID: spec.AccountID,
		Role:        spec.Role,
		Issuer:      s.domain,
		IssuedAt:    now,
		ExpiresAt:   now.Add(spec.TTL),
		SourceType:  spec.SourceType,
		SourceID:    spec.SourceID,
	}
	if err := account.PersistToken(tx, &token, spec.Limits); err != nil {
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

func lockGrantSource(tx *gorm.DB, accountID, tokenID string, historical bool) (*app.Token, error) {
	if accountID == "" || tokenID == "" {
		return nil, errTokenNotFound
	}
	var account app.Account
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id", "account_type").
		Where(app.Account{ID: accountID}).
		First(&account).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errTokenNotFound
	}
	if err != nil {
		return nil, err
	}
	if account.AccountType != app.AccountTypeAuth && account.AccountType != app.AccountTypeAuth0 {
		return nil, errTokenNotFound
	}
	if historical {
		tx = tx.Unscoped()
	}

	var token app.Token
	err = tx.Clauses(clause.Locking{Strength: "SHARE"}).
		Where(app.Token{ID: tokenID, AccountID: accountID, TokenType: app.TokenTypeNuon}).
		First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errTokenNotFound
	}
	if err != nil {
		return nil, err
	}
	limits, err := permissions.DecodeCredentialLimits(token.Authorization)
	if err != nil || limits.Profile != permissions.ProfileHuman || limits.API != permissions.APICtl || token.IssuedAt.After(time.Now()) {
		return nil, errTokenNotFound
	}
	if !historical && !time.Now().Before(token.ExpiresAt) {
		return nil, errTokenExpired
	}
	return &token, nil
}

func (s *service) createToken(db *gorm.DB, account *app.Account) (string, error) {
	var token *app.Token
	err := db.Transaction(func(tx *gorm.DB) error {
		limits, err := authz.New(authz.Params{DB: tx}).HumanCredentialLimits(tx.Statement.Context, account.ID)
		if err != nil {
			return err
		}
		token, err = s.createAccessToken(tx, accessToken{
			AccountID: account.ID, TokenType: app.TokenTypeNuon,
			TTL: time.Duration(s.cfg.NuonAuthTokenTTL) * time.Minute, Limits: limits,
		})
		return err
	})
	if err != nil {
		return "", err
	}

	return token.Token, nil
}

func grantAuthorization(tx *gorm.DB, accountID, tokenID string, allowed permissions.PermissionSet) ([]byte, error) {
	source, err := lockGrantSource(tx, accountID, tokenID, false)
	if err != nil {
		return nil, err
	}
	limits, err := permissions.DecodeCredentialLimits(source.Authorization)
	if err != nil {
		return nil, err
	}
	limits, err = authz.New(authz.Params{DB: tx}).ConstrainHumanCredential(tx.Statement.Context, accountID, limits)
	if err != nil {
		return nil, err
	}
	if allowed != nil {
		for _, grant := range limits.Ceiling {
			for permission := range grant.Permissions {
				if _, ok := allowed[permission]; !ok {
					delete(grant.Permissions, permission)
				}
			}
		}
	}
	return limits.MarshalJSON()
}

func constrainGrantAuthorization(tx *gorm.DB, accountID string, authorization []byte) (permissions.CredentialLimits, error) {
	limits, err := permissions.DecodeCredentialLimits(authorization)
	if err != nil || limits.Profile != permissions.ProfileHuman || limits.API != permissions.APICtl {
		return permissions.CredentialLimits{}, errGrantAuthorization
	}
	return authz.New(authz.Params{DB: tx}).ConstrainHumanCredential(tx.Statement.Context, accountID, limits)
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
	if !time.Now().Before(token.ExpiresAt) || token.IssuedAt.After(time.Now()) {
		return nil, errTokenExpired
	}
	limits, err := permissions.DecodeCredentialLimits(token.Authorization)
	if err != nil || limits.Profile != permissions.ProfileHuman || limits.API != permissions.APICtl {
		return nil, errTokenNotFound
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
	if account.AccountType != app.AccountTypeAuth && account.AccountType != app.AccountTypeAuth0 {
		return nil, errTokenNotFound
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
