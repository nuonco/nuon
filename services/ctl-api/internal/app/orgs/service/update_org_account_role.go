package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type UpdateOrgAccountRoleRequest struct {
	RoleType app.RoleType `json:"role_type" validate:"required"`
}

var allowedAccountRoles = map[app.RoleType]struct{}{
	app.RoleTypeOrgAdmin:    {},
	app.RoleTypeOrgReadOnly: {},
}

// @ID						UpdateOrgAccountRole
// @Summary				Change an org member's role
// @Description			Changes the role of an existing member of the current org. Requires org admin. You cannot change your own role, and you cannot demote the last remaining admin.
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
func (s *service) UpdateOrgAccountRole(ctx *gin.Context) {
	accountID := ctx.Param("account_id")

	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	caller, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	if !s.isOrgAdmin(caller, org.ID) {
		ctx.Error(stderr.ErrAuthorization{
			Err:         fmt.Errorf("only org admins can change member roles"),
			Description: "only org admins can change member roles",
		})
		return
	}

	var req UpdateOrgAccountRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse request: %w", err),
			Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
		})
		return
	}

	if _, ok := allowedAccountRoles[req.RoleType]; !ok {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid role type: %s", req.RoleType),
			Description: fmt.Sprintf("role_type must be %q or %q", app.RoleTypeOrgAdmin, app.RoleTypeOrgReadOnly),
		})
		return
	}

	if accountID == caller.ID {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("cannot change your own role"),
			Description: "you cannot change your own role",
		})
		return
	}

	target, err := s.getOrgAccount(ctx, org.ID, accountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.Error(stderr.ErrNotFound{Err: fmt.Errorf("account not found in org")})
			return
		}
		ctx.Error(fmt.Errorf("unable to look up account: %w", err))
		return
	}

	if s.isOrgAdmin(target, org.ID) && req.RoleType != app.RoleTypeOrgAdmin {
		count, err := s.countOrgAdmins(ctx, org.ID)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to count org admins: %w", err))
			return
		}
		if count <= 1 {
			ctx.Error(stderr.ErrUser{
				Err:         fmt.Errorf("cannot demote the last org admin"),
				Description: "the org must have at least one admin; promote another member before changing this role",
			})
			return
		}
	}

	if err := s.authzClient.SetAccountOrgRole(ctx, org.ID, accountID, req.RoleType); err != nil {
		ctx.Error(fmt.Errorf("unable to change member role: %w", err))
		return
	}

	updated, err := s.getOrgAccount(ctx, org.ID, accountID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to reload account: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, updated)
}

func (s *service) getOrgAccount(ctx context.Context, orgID, accountID string) (*app.Account, error) {
	var account app.Account
	res := s.db.WithContext(ctx).
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
	var count int64
	res := s.db.WithContext(ctx).
		Model(&app.AccountRole{}).
		Joins("JOIN roles ON roles.id = account_roles.role_id").
		Where("account_roles.org_id = ? AND account_roles.deleted_at = 0", orgID).
		Where("roles.role_type = ?", app.RoleTypeOrgAdmin).
		Distinct("account_roles.account_id").
		Count(&count)
	return count, res.Error
}
