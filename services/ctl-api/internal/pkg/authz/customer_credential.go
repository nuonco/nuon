package authz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"gorm.io/gorm"
)

func CustomerCredential(account *app.Account, token *app.Token, surface cctx.CredentialSurface) (cctx.VerifiedCredential, error) {
	if account == nil || token == nil {
		return cctx.VerifiedCredential{}, fmt.Errorf("account and token are required")
	}
	credential := cctx.VerifiedCredential{AccountID: account.ID, TokenID: token.ID, Surface: surface, Audience: permissions.APICtl}
	switch account.AccountType {
	case app.AccountTypeAuth, app.AccountTypeAuth0:
		credential.Profile = cctx.ProfileHuman
	case app.AccountTypeService:
		credential.Profile = cctx.ProfileServiceAccount
	}
	if !customerCredentialAllowed(credential) || account.DeletedAt != 0 || token.DeletedAt != 0 {
		return cctx.VerifiedCredential{}, fmt.Errorf("customer credential profile or lifecycle is not allowed")
	}
	ceiling, reason := customerTokenCeiling(credential, *account, *token, token.OrgID, time.Now())
	if reason != ReasonAllowed {
		return cctx.VerifiedCredential{}, fmt.Errorf("customer credential denied: %s", reason)
	}
	credential.Ceiling = ceiling
	return credential, nil
}

func (c *Client) PrepareCustomer(ctx context.Context, orgID string) (context.Context, error) {
	credential, ok := cctx.VerifiedCredentialFromContext(ctx)
	if !ok || !customerCredentialAllowed(credential) {
		return ctx, fmt.Errorf("verified customer credential required")
	}
	_, reason, err := loadCustomerCredential(c.db.WithContext(ctx), credential, orgID, time.Now())
	if err != nil {
		return ctx, err
	}
	if reason != ReasonAllowed {
		return ctx, fmt.Errorf("customer credential denied: %s", reason)
	}
	return ctx, nil
}

func loadCustomerCredential(db *gorm.DB, credential cctx.VerifiedCredential, orgID string, now time.Time) ([]permissions.ScopedPermissions, DecisionReason, error) {
	var token app.Token
	if err := db.Select("id", "account_id", "org_id", "token_type", "issued_at", "expires_at", "role", "source_type", "source_id", "authorization").
		Where(app.Token{ID: credential.TokenID, AccountID: credential.AccountID}).Take(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ReasonCredentialUnavailable, nil
		}
		return nil, ReasonCredentialUnavailable, fmt.Errorf("load authorization credential: %w", err)
	}
	var account app.Account
	if err := db.Select("id", "account_type").Where(app.Account{ID: credential.AccountID}).Take(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ReasonCredentialUnavailable, nil
		}
		return nil, ReasonCredentialUnavailable, fmt.Errorf("load authorization principal: %w", err)
	}
	ceiling, reason := customerTokenCeiling(credential, account, token, orgID, now)
	if reason != ReasonAllowed {
		return nil, reason, nil
	}
	if credential.Profile == cctx.ProfileServiceAccount {
		var managed app.ManagedServiceAccount
		err := db.Unscoped().Select("account_id").Where(app.ManagedServiceAccount{AccountID: account.ID}).Take(&managed).Error
		if err == nil {
			return nil, ReasonCredentialNotAllowed, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ReasonCredentialUnavailable, fmt.Errorf("load authorization account ownership: %w", err)
		}
	}
	return ceiling, ReasonAllowed, nil
}

func customerCredentialAllowed(credential cctx.VerifiedCredential) bool {
	return credential.AccountID != "" && credential.TokenID != "" && credential.Audience == permissions.APICtl &&
		(credential.Surface == cctx.SurfacePublicAPI || credential.Surface == cctx.SurfaceMCP) &&
		(credential.Profile == cctx.ProfileHuman || (credential.Profile == cctx.ProfileServiceAccount && credential.Surface == cctx.SurfacePublicAPI))
}

func customerTokenCeiling(credential cctx.VerifiedCredential, account app.Account, token app.Token, orgID string, now time.Time) ([]permissions.ScopedPermissions, DecisionReason) {
	deny := func(reason DecisionReason) ([]permissions.ScopedPermissions, DecisionReason) {
		return nil, reason
	}
	if !now.Before(token.ExpiresAt) {
		return deny(ReasonCredentialExpired)
	}
	if token.IssuedAt.After(now) || token.AccountID != account.ID {
		return deny(ReasonCredentialNotAllowed)
	}
	if token.OrgID != "" && token.OrgID != orgID {
		return deny(ReasonTenantMismatch)
	}
	switch credential.Profile {
	case cctx.ProfileHuman:
		if account.AccountType != app.AccountTypeAuth && account.AccountType != app.AccountTypeAuth0 {
			return deny(ReasonCredentialNotAllowed)
		}
	case cctx.ProfileServiceAccount:
		if account.AccountType != app.AccountTypeService || credential.Surface != cctx.SurfacePublicAPI {
			return deny(ReasonCredentialNotAllowed)
		}
	default:
		return deny(ReasonCredentialNotAllowed)
	}
	if credential.Surface == cctx.SurfaceMCP && token.TokenType == app.TokenTypeStatic {
		return deny(ReasonCredentialNotAllowed)
	}
	switch token.TokenType {
	case app.TokenTypeNuon, app.TokenTypeAuth, app.TokenTypeAuth0:
		if token.TokenType != app.TokenTypeNuon && account.AccountType == app.AccountTypeService {
			return deny(ReasonCredentialNotAllowed)
		}
	case app.TokenTypeStatic, app.TokenTypeOAuth, app.TokenTypeFederated:
		if token.TokenType == app.TokenTypeOAuth && account.AccountType == app.AccountTypeService {
			return deny(ReasonCredentialNotAllowed)
		}
		if account.AccountType == app.AccountTypeService && token.OrgID != orgID {
			return deny(ReasonTenantMismatch)
		}
		if token.TokenType == app.TokenTypeFederated && (account.AccountType != app.AccountTypeService ||
			token.SourceType != app.TokenSourceTypeOIDCTrustPolicy || token.SourceID == "") {
			return deny(ReasonCredentialNotAllowed)
		}
	default:
		return deny(ReasonCredentialNotAllowed)
	}
	limits, err := permissions.DecodeCredentialLimits(token.Authorization)
	if err != nil || limits.Profile != credential.Profile || limits.API != credential.Audience {
		return deny(ReasonCredentialNotAllowed)
	}
	for _, grant := range limits.Ceiling {
		if token.OrgID != "" && grant.Scope.OrgID != token.OrgID {
			return deny(ReasonTenantMismatch)
		}
	}
	return limits.Ceiling, ReasonAllowed
}
