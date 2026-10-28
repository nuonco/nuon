package helpers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const DefaultTokenDuration = "8760h"

// IsOrgAdmin reports whether acct has the org-admin role for orgID.
func (h *Helpers) IsOrgAdmin(acct *app.Account, orgID string) bool {
	for _, role := range acct.Roles {
		if role.RoleType == app.RoleTypeOrgAdmin && role.OrgID.ValueString() == orgID {
			return true
		}
	}
	return false
}

// RequireOrgAdmin loads the request's org and account and ensures the account
// is an org admin for that org.
func (h *Helpers) RequireOrgAdmin(ctx *gin.Context) (*app.Org, error) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		return nil, err
	}

	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		return nil, err
	}

	if !h.IsOrgAdmin(acct, org.ID) {
		return nil, stderr.ErrAuthorization{
			Err:         fmt.Errorf("only org admins can manage static tokens"),
			Description: "only org admins can manage static tokens",
		}
	}

	return org, nil
}

// OrgRoleType reports the account's org role for display only.
func OrgRoleType(acct *app.Account, orgID string) app.RoleType {
	for _, role := range acct.Roles {
		if role.OrgID.ValueString() == orgID {
			return role.RoleType
		}
	}

	return ""
}

// GetOrgServiceAccount looks up an account by ID and ensures it is a service
// account that belongs to the given org.
func (h *Helpers) GetOrgServiceAccount(ctx context.Context, orgID, accountID string) (*app.Account, error) {
	acct, err := h.acctClient.FindAccount(ctx, accountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, stderr.ErrNotFound{
				Err:         fmt.Errorf("service account %q not found", accountID),
				Description: "service account not found",
			}
		}
		return nil, fmt.Errorf("unable to find account: %w", err)
	}

	if acct.AccountType != app.AccountTypeService {
		return nil, stderr.ErrNotFound{
			Err:         fmt.Errorf("account %q is not a service account", accountID),
			Description: "service account not found",
		}
	}

	found := false
	for _, orgIDValue := range acct.OrgIDs {
		if orgIDValue == orgID {
			found = true
			break
		}
	}
	if !found {
		return nil, stderr.ErrAuthorization{
			Err:         fmt.Errorf("service account %q does not belong to org %q", accountID, orgID),
			Description: "service account not found for this org",
		}
	}

	return acct, nil
}

// OrgServiceAccountIDs resolves the org's service accounts off the
// account_roles org index. Joining accounts to account_roles in the list query
// instead lets the planner walk the whole accounts table in email order to
// satisfy the ORDER BY and LIMIT, which took tens of seconds on a cold cache.
func (h *Helpers) OrgServiceAccountIDs(ctx context.Context, orgID string, includeRunners, includeStacks bool) ([]string, error) {
	accountIDs := []string{}
	if err := h.OrgServiceAccounts(ctx, orgID, includeRunners, includeStacks).Pluck("account_roles.account_id", &accountIDs).Error; err != nil {
		return nil, fmt.Errorf("unable to list service account ids for org %s: %w", orgID, err)
	}

	return accountIDs, nil
}

func (h *Helpers) OrgServiceAccounts(ctx context.Context, orgID string, includeRunners, includeStacks bool) *gorm.DB {
	tx := h.db.WithContext(ctx).
		Model(&app.AccountRole{}).
		Joins("JOIN accounts ON accounts.id = account_roles.account_id AND accounts.deleted_at = 0 AND accounts.account_type = ?", app.AccountTypeService).
		Where(app.AccountRole{OrgID: generics.NewNullString(orgID)})

	excludedRoleTypes := []app.RoleType{}
	if !includeRunners {
		excludedRoleTypes = append(excludedRoleTypes, app.RoleTypeRunner)
	}
	if !includeStacks {
		excludedRoleTypes = append(excludedRoleTypes, app.RoleTypeStack)
	}
	if len(excludedRoleTypes) > 0 {
		tx = tx.
			Joins("JOIN roles ON roles.id = account_roles.role_id AND roles.deleted_at = 0").
			Where("roles.role_type NOT IN ?", excludedRoleTypes)
	}

	return tx.Distinct("account_roles.account_id")
}

// CreateStaticToken issues a static API token for acct.
func (h *Helpers) CreateStaticToken(ctx context.Context, acct *app.Account, orgID, createdByID, name string, role app.RoleType, duration time.Duration) (*app.Token, error) {
	token := app.Token{
		CreatedByID: createdByID,
		Name:        name,
		Role:        string(role),
		OrgID:       orgID,
		Token:       domains.NewUserTokenID(),
		TokenType:   app.TokenTypeStatic,
		ExpiresAt:   time.Now().Add(duration),
		IssuedAt:    time.Now(),
		Issuer:      acct.ID,
		AccountID:   acct.ID,
	}

	if err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var liveAccount struct{ ID string }
		if err := tx.Model(&app.Account{}).Select("id").Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(app.Account{ID: acct.ID}).Take(&liveAccount).Error; err != nil {
			return fmt.Errorf("lock account for token issuance: %w", err)
		}
		return tx.Create(&token).Error
	}); err != nil {
		return nil, fmt.Errorf("unable to create static token: %w", err)
	}

	return &token, nil
}

// CreateServiceAccountToken creates a static token for the org's service account
// identified by accountID, optionally invalidating its existing tokens first.
func (h *Helpers) CreateServiceAccountToken(ctx *gin.Context, org *app.Org, accountID, name, duration string, invalidate bool) (*app.Token, error) {
	if duration == "" {
		duration = DefaultTokenDuration
	}

	dur, err := time.ParseDuration(duration)
	if err != nil {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("invalid duration: %w", err),
			Description: fmt.Sprintf("invalid duration: %s", err.Error()),
		}
	}

	acct, err := h.GetOrgServiceAccount(ctx, org.ID, accountID)
	if err != nil {
		return nil, err
	}

	if invalidate {
		if err := h.acctClient.InvalidateTokens(ctx, acct.Email); err != nil {
			return nil, fmt.Errorf("unable to invalidate tokens: %w", err)
		}
	}

	caller, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		return nil, err
	}

	name = strings.TrimSpace(name)
	if name == "" {
		name = acct.Email
	}

	// CreateStaticToken, not acctClient.CreateToken: only this one stamps the columns
	// the token list endpoints filter on, so tokens made the other way cannot be
	// listed or revoked.
	token, err := h.CreateStaticToken(ctx, acct, org.ID, caller.ID, name, OrgRoleType(acct, org.ID), dur)
	if err != nil {
		return nil, fmt.Errorf("unable to create token: %w", err)
	}

	return token, nil
}
