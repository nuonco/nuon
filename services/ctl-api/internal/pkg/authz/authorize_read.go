package authz

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

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
	case permissions.OrgRead, permissions.OrgWrite, permissions.AppRead, permissions.InstallRead, permissions.TelemetryRead,
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
	if !customerCredentialAllowed(credential) {
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
		decision, err = authorize(tx, credential, req.Permission, req.Target, time.Now())
		return err
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return Decision{}, err
	}
	return decision, nil
}

func authorize(db *gorm.DB, credential cctx.VerifiedCredential, permission permissions.PermissionID, scope permissions.Scope, now time.Time) (Decision, error) {
	ceiling, reason, err := loadCustomerCredential(db, credential, scope.OrgID, now)
	if err != nil || reason != ReasonAllowed {
		return Decision{Reason: reason}, err
	}
	target, err := resolveTarget(db, scope)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Decision{Reason: ReasonTargetUnavailable}, nil
		}
		return Decision{}, err
	}
	grants, err := loadScopedGrants(db, credential.AccountID, scope.OrgID)
	if err != nil {
		if errors.Is(err, errNoScopedGrants) {
			return Decision{Reason: ReasonNoGrant}, nil
		}
		return Decision{}, err
	}
	if grants.API != credential.Audience {
		return Decision{Reason: ReasonCredentialNotAllowed}, nil
	}
	if match := permissions.MatchPermissions(permissions.PermissionCheck{
		Permission: permission, Target: target, Grants: grants.Grants, Ceiling: ceiling,
	}); !match.Matched {
		if match.Reason == permissions.ReasonNoGrant {
			return Decision{Reason: ReasonNoGrant}, nil
		}
		return Decision{Reason: ReasonCeilingExceeded}, nil
	}
	match := permissions.MatchPermissions(permissions.PermissionCheck{
		Permission: permission, Target: target, Grants: grants.Grants, Ceiling: credential.Ceiling,
	})
	switch match.Reason {
	case permissions.ReasonMatched:
		return Decision{Allowed: true, Reason: ReasonAllowed, Target: target}, nil
	case permissions.ReasonNoGrant:
		return Decision{Reason: ReasonNoGrant}, nil
	case permissions.ReasonCeilingExceeded:
		return Decision{Reason: ReasonCeilingExceeded}, nil
	default:
		return Decision{}, fmt.Errorf("invalid authorization inputs: %s", match.Reason)
	}
}
