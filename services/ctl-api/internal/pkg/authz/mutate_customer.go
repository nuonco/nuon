package authz

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type MutationRequest struct {
	Permission       permissions.PermissionID
	Target           permissions.Scope
	TargetReference  string
	SubjectAccountID string
}

func (c *Client) MutateCustomer(ctx context.Context, req MutationRequest, mutate func(*gorm.DB, permissions.ResolvedTarget) error) (Decision, error) {
	switch req.Permission {
	case permissions.OrgWrite, permissions.AppWrite, permissions.InstallWrite, permissions.AppDelete, permissions.InstallDelete:
	default:
		return Decision{Reason: ReasonInvalidRequest}, nil
	}
	if req.TargetReference != "" {
		if req.Target.ID != "" {
			return Decision{Reason: ReasonInvalidRequest}, nil
		}
		req.Target.ID = req.TargetReference
	}
	if req.Target.ValidateShape() != nil || mutate == nil ||
		!req.Permission.AppliesTo(req.Target.Kind) {
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
		if err := requireRoleMutationIsolation(tx); err != nil {
			return err
		}
		decision.Reason = ReasonTargetUnavailable
		if err := lockRoleOrg(tx, req.Target.OrgID); err != nil {
			return err
		}
		if req.TargetReference != "" {
			var err error
			req.Target, err = resolveReference(tx, req.Target.OrgID, req.Target.Kind, req.TargetReference, nil)
			if err != nil {
				return err
			}
		}
		decision.Reason = ReasonCredentialUnavailable
		accountIDs := []string{credential.AccountID}
		if req.SubjectAccountID != "" && req.SubjectAccountID != credential.AccountID {
			accountIDs = append(accountIDs, req.SubjectAccountID)
		}
		slices.Sort(accountIDs)
		for _, accountID := range accountIDs {
			if _, err := lockRoleSubject(tx, accountID); err != nil {
				return err
			}
		}
		var token app.Token
		if err := tx.Select("id").Clauses(clause.Locking{Strength: "SHARE"}).
			Where(app.Token{ID: credential.TokenID, AccountID: credential.AccountID}).Take(&token).Error; err != nil {
			return err
		}
		decision.Reason = ReasonNoGrant
		var bindings []app.AccountRole
		if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).
			Preload("Role", func(db *gorm.DB) *gorm.DB {
				return db.Clauses(clause.Locking{Strength: "SHARE"}).Order("id")
			}).Preload("Role.Policies", func(db *gorm.DB) *gorm.DB {
			return db.Clauses(clause.Locking{Strength: "SHARE"}).Order("id")
		}).Where(app.AccountRole{AccountID: credential.AccountID, OrgID: generics.NewNullString(req.Target.OrgID)}).Order("id").Find(&bindings).Error; err != nil {
			return err
		}
		decision.Reason = ReasonTargetUnavailable
		if _, err := resolveTarget(tx.Clauses(clause.Locking{Strength: "UPDATE"}).Session(&gorm.Session{}), req.Target); err != nil {
			return err
		}
		var err error
		decision, err = authorize(tx, credential, req.Permission, req.Target, time.Now())
		if err != nil || !decision.Allowed {
			return err
		}
		return mutate(tx, decision.Target)
	}, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		if !decision.Allowed && errors.Is(err, gorm.ErrRecordNotFound) {
			return decision, nil
		}
		return Decision{}, err
	}
	return decision, nil
}
