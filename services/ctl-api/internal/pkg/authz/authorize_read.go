package authz

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type ReadRequest struct {
	Permission permissions.PermissionID
	Target     permissions.Scope
}

type DecisionReason string

const (
	ReasonAllowed               DecisionReason = "allowed"
	ReasonUnsupportedRead       DecisionReason = "unsupported_read"
	ReasonInvalidRequest        DecisionReason = "invalid_request"
	ReasonMissingCredential     DecisionReason = "missing_credential"
	ReasonCredentialNotAllowed  DecisionReason = "credential_not_allowed"
	ReasonCredentialUnavailable DecisionReason = "credential_unavailable"
	ReasonCredentialExpired     DecisionReason = "credential_expired"
	ReasonTenantMismatch        DecisionReason = "tenant_mismatch"
	ReasonTargetUnavailable     DecisionReason = "target_unavailable"
	ReasonNoGrant               DecisionReason = "no_grant"
	ReasonCeilingExceeded       DecisionReason = "ceiling_exceeded"
)

type Decision struct {
	Allowed bool
	Reason  DecisionReason
	Target  permissions.ResolvedTarget
}

func (c *Client) AuthorizeRead(ctx context.Context, req ReadRequest) (Decision, error) {
	switch req.Permission {
	case permissions.OrgRead, permissions.AppRead, permissions.InstallRead, permissions.TelemetryRead,
		permissions.RunnerRead, permissions.RunnerJobRead, permissions.LogStreamRead:
	default:
		return Decision{Reason: ReasonUnsupportedRead}, nil
	}
	if req.Target.ValidateShape() != nil || !req.Permission.AppliesTo(req.Target.Kind) {
		return Decision{Reason: ReasonInvalidRequest}, nil
	}
	credential, ok := cctx.VerifiedCredentialFromContext(ctx)
	if !ok || credential.AccountID == "" || credential.TokenID == "" {
		return Decision{Reason: ReasonMissingCredential}, nil
	}
	if credential.Audience != permissions.APICtl ||
		(credential.Surface != cctx.SurfacePublicAPI && credential.Surface != cctx.SurfaceMCP) ||
		(credential.Profile != cctx.ProfileHuman && credential.Profile != cctx.ProfileServiceAccount) {
		return Decision{Reason: ReasonCredentialNotAllowed}, nil
	}
	if len(credential.Ceiling) == 0 {
		return Decision{Reason: ReasonCeilingExceeded}, nil
	}
	var decision Decision
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var isolation string
		if err := tx.Raw("SHOW transaction_isolation").Scan(&isolation).Error; err != nil {
			return err
		}
		if isolation != "repeatable read" && isolation != "serializable" {
			return fmt.Errorf("read authorization requires repeatable-read or serializable isolation")
		}
		var err error
		decision, err = authorizeRead(tx, credential, req, time.Now())
		return err
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return Decision{}, err
	}
	return decision, nil
}

func authorizeRead(db *gorm.DB, credential cctx.VerifiedCredential, req ReadRequest, now time.Time) (Decision, error) {
	var token app.Token
	if err := db.Select("id", "account_id", "org_id", "token_type", "issued_at", "expires_at").
		Where(app.Token{ID: credential.TokenID, AccountID: credential.AccountID}).Take(&token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Decision{Reason: ReasonCredentialUnavailable}, nil
		}
		return Decision{}, fmt.Errorf("load authorization credential: %w", err)
	}
	if !now.Before(token.ExpiresAt) {
		return Decision{Reason: ReasonCredentialExpired}, nil
	}
	if token.IssuedAt.After(now) {
		return Decision{Reason: ReasonCredentialNotAllowed}, nil
	}
	if token.OrgID != "" && token.OrgID != req.Target.OrgID {
		return Decision{Reason: ReasonTenantMismatch}, nil
	}
	var account app.Account
	if err := db.Select("id", "account_type").Where(app.Account{ID: credential.AccountID}).Take(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Decision{Reason: ReasonCredentialUnavailable}, nil
		}
		return Decision{}, fmt.Errorf("load authorization principal: %w", err)
	}
	switch credential.Profile {
	case cctx.ProfileHuman:
		if (account.AccountType != app.AccountTypeAuth && account.AccountType != app.AccountTypeAuth0) ||
			(token.TokenType != app.TokenTypeNuon && token.TokenType != app.TokenTypeOAuth) {
			return Decision{Reason: ReasonCredentialNotAllowed}, nil
		}
	case cctx.ProfileServiceAccount:
		if account.AccountType != app.AccountTypeService || token.TokenType != app.TokenTypeStatic ||
			credential.Surface != cctx.SurfacePublicAPI || token.OrgID != req.Target.OrgID {
			return Decision{Reason: ReasonCredentialNotAllowed}, nil
		}
		var managed app.ManagedServiceAccount
		err := db.Unscoped().Select("account_id").Where(app.ManagedServiceAccount{AccountID: account.ID}).Take(&managed).Error
		if err == nil {
			return Decision{Reason: ReasonCredentialNotAllowed}, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return Decision{}, fmt.Errorf("load authorization account ownership: %w", err)
		}
	}
	target, err := resolveTarget(db, req.Target)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Decision{Reason: ReasonTargetUnavailable}, nil
		}
		return Decision{}, err
	}
	grants, err := loadScopedGrants(db, account.ID, req.Target.OrgID)
	if err != nil {
		if errors.Is(err, errNoScopedGrants) {
			return Decision{Reason: ReasonNoGrant}, nil
		}
		return Decision{}, err
	}
	if grants.API != credential.Audience {
		return Decision{Reason: ReasonCredentialNotAllowed}, nil
	}
	match := permissions.MatchPermissions(permissions.PermissionCheck{
		Permission: req.Permission, Target: target, Grants: grants.Grants, Ceiling: credential.Ceiling,
	})
	switch match.Reason {
	case permissions.ReasonMatched:
		return Decision{Allowed: true, Reason: ReasonAllowed, Target: target}, nil
	case permissions.ReasonNoGrant:
		return Decision{Reason: ReasonNoGrant}, nil
	case permissions.ReasonCeilingExceeded:
		return Decision{Reason: ReasonCeilingExceeded}, nil
	default:
		return Decision{}, fmt.Errorf("invalid read authorization inputs: %s", match.Reason)
	}
}
