package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type UpdateOrgAccountRoleRequest struct {
	RoleType app.RoleType       `json:"role_type" binding:"required"`
	Scope    *permissions.Scope `json:"scope,omitempty"`
	Remove   bool               `json:"remove,omitempty"`
}

// @ID						UpdateOrgAccountRole
// @Summary				Change an org member's role
// @Description			Replaces bindings at one scope, preserving other scopes. Omit scope for the current org. Set remove=true with an explicit scope and role_type to remove only that binding. Removal requires org.write; assignment also requires delegable permissions. You cannot change your own role, remove the final binding or demote the last org admin. Responses include current-org role_bindings.
// @Param					account_id	path	string						true	"account ID"
// @Param					req			body	UpdateOrgAccountRoleRequest	true	"Input"
// @Tags					orgs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Success				200	{object}	app.Account
// @Router					/v1/orgs/current/accounts/{account_id}/role [PATCH]
func (s *service) UpdateOrgAccountRole(ctx *gin.Context) (require.Response, error) {
	access, err := require.MutationAccessFromContext(ctx)
	if err != nil {
		return require.Response{}, err
	}
	accountID := access.SubjectAccountID()
	orgID := access.Target().OrgID
	tx := access.DB()

	caller, err := cctx.AccountFromContext(tx.Statement.Context)
	if err != nil {
		return require.Response{}, err
	}

	var req UpdateOrgAccountRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return require.Response{}, stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse request: %w", err),
			Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
		}
	}

	if req.Remove && req.Scope == nil {
		return require.Response{}, stderr.NewInvalidRequest(fmt.Errorf("removal requires an explicit scope"))
	}
	bindingScope, err := authz.RoleBindingScope(orgID, req.Scope)
	if err != nil {
		return require.Response{}, stderr.NewInvalidRequest(err)
	}
	if accountID == caller.ID {
		return require.Response{}, stderr.ErrUser{Err: fmt.Errorf("cannot change your own role"), Description: "you cannot change your own role"}
	}
	target, err := prepareOrgMember(tx, orgID, accountID, true)
	if err != nil {
		return require.Response{}, err
	}
	if bindingScope.Kind == permissions.KindOrg && ((!req.Remove && req.RoleType != app.RoleTypeOrgAdmin) || (req.Remove && req.RoleType == app.RoleTypeOrgAdmin)) {
		if err := s.requireRemainingOrgAdmin(tx, orgID, target); err != nil {
			return require.Response{}, err
		}
	}
	client := authz.New(authz.Params{DB: tx})
	if req.Remove {
		role, err := client.ResolveAssignableRole(tx.Statement.Context, orgID, req.RoleType, app.RoleContextTeam)
		if err != nil {
			return require.Response{}, err
		}
		if err := client.RemoveAccountRoleAtScope(tx.Statement.Context, accountID, role.ID, bindingScope); err != nil {
			return require.Response{}, err
		}
	} else {
		role, err := client.ResolveDelegableRoleAtScope(tx.Statement.Context, bindingScope, req.RoleType, app.RoleContextTeam)
		if err != nil {
			return require.Response{}, err
		}
		if err := client.SetAccountRoleAtScope(tx.Statement.Context, accountID, role.ID, bindingScope); err != nil {
			return require.Response{}, err
		}
	}

	updated, err := getOrgAccount(tx, orgID, accountID)
	if err != nil {
		return require.Response{}, fmt.Errorf("unable to reload account: %w", err)
	}
	bindings, err := client.AccountRoleBindings(tx.Statement.Context, orgID, []string{accountID})
	if err != nil {
		return require.Response{}, err
	}
	updated.RoleBindings = bindings[accountID]

	return require.Response{Status: http.StatusOK, Body: updated}, nil
}

func (s *service) getOrgAccount(ctx context.Context, orgID, accountID string) (*app.Account, error) {
	return getOrgAccount(s.db.WithContext(ctx), orgID, accountID)
}

func getOrgAccount(db *gorm.DB, orgID, accountID string) (*app.Account, error) {
	var account app.Account
	res := db.
		Joins("JOIN account_roles ON account_roles.account_id = accounts.id AND account_roles.org_id = ? AND account_roles.deleted_at = 0", orgID).
		Where("accounts.id = ?", accountID).
		Group("accounts.id").
		Preload("Roles", "org_id = ?", orgID).
		Preload("Roles.Org").
		Preload("Roles.Policies").
		First(&account)
	if res.Error != nil {
		return nil, res.Error
	}
	return &account, nil
}

func (s *service) countOrgAdmins(ctx context.Context, orgID string) (int64, error) {
	return countOrgAdmins(s.db.WithContext(ctx), orgID)
}

func countOrgAdmins(db *gorm.DB, orgID string) (int64, error) {
	var count int64
	res := db.
		Model(&app.AccountRole{}).
		Joins("JOIN roles ON roles.id = account_roles.role_id AND roles.deleted_at = 0").
		Joins("JOIN accounts ON accounts.id = account_roles.account_id AND accounts.deleted_at = 0").
		Where("account_roles.org_id = ? AND account_roles.deleted_at = 0", orgID).
		Where("roles.org_id = ?", orgID).
		Where("roles.role_type = ?", app.RoleTypeOrgAdmin).
		Where("roles.version = ?", app.RoleVersionV3).
		Where(app.AccountRole{ScopeKind: permissions.KindOrg, ScopeID: orgID}).
		Distinct("account_roles.account_id").
		Count(&count)
	return count, res.Error
}

func prepareOrgMember(tx *gorm.DB, orgID, accountID string, existing bool) (*app.Account, error) {
	if err := authz.RequireUserManagedInOrg(tx, accountID, orgID); err != nil {
		return nil, err
	}
	var account app.Account
	if err := tx.Where(app.Account{ID: accountID}).Take(&account).Error; err != nil {
		return nil, err
	}
	if account.AccountType != app.AccountTypeAuth && account.AccountType != app.AccountTypeAuth0 {
		return nil, stderr.ErrUser{Err: fmt.Errorf("team membership requires a human account"), Description: "manage service accounts through the service-account API"}
	}
	if existing {
		member, err := getOrgAccount(tx, orgID, accountID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, stderr.ErrNotFound{Err: fmt.Errorf("account not found in org")}
			}
			return nil, err
		}
		return member, nil
	}
	var count int64
	if err := tx.Model(&app.AccountRole{}).Where(app.AccountRole{AccountID: accountID, OrgID: generics.NewNullString(orgID)}).Count(&count).Error; err != nil {
		return nil, err
	}
	if count != 0 {
		return nil, stderr.ErrUser{Err: fmt.Errorf("account is already a member"), Description: "account is already a member"}
	}
	return &account, nil
}

func (s *service) requireRemainingOrgAdmin(tx *gorm.DB, orgID string, target *app.Account) error {
	if !s.isOrgAdmin(target, orgID) {
		return nil
	}
	count, err := countOrgAdmins(tx, orgID)
	if err != nil {
		return err
	}
	if count <= 1 {
		return stderr.ErrUser{Err: fmt.Errorf("cannot remove the last org admin"), Description: "the org must have at least one admin; promote another member first"}
	}
	return nil
}
