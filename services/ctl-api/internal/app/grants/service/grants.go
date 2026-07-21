package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// CreateGrantRequest grants an account a permission on a resource. Exactly one
// of AccountID or Email identifies the grantee.
type CreateGrantRequest struct {
	AccountID  string `json:"account_id"`
	Email      string `json:"email"`
	Permission string `json:"permission" validate:"required,oneof=read all"`
}

// CreateInstallGrant godoc
//
//	@ID				CreateInstallGrant
//	@Summary		grant an account access to an install
//	@Description	Grant an account read or full access to a single install. Org-admin only.
//	@Tags			grants
//	@Accept			json
//	@Produce		json
//	@Security		APIKey
//	@Security		OrgID
//	@Param			install_id	path		string				true	"install ID"
//	@Param			req			body		CreateGrantRequest	true	"grant"
//	@Success		201			{object}	app.ResourceGrant
//	@Failure		400			{object}	stderr.ErrResponse
//	@Failure		403			{object}	stderr.ErrResponse
//	@Router			/v1/installs/{install_id}/grants [POST]
func (s *service) CreateInstallGrant(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.requireOrgAdmin(ctx, org.ID); err != nil {
		ctx.Error(err)
		return
	}

	inst := app.Install{}
	if err := s.resolveByNameOrID(ctx, org.ID, ctx.Param("install_id"), &inst); err != nil {
		ctx.Error(fmt.Errorf("unable to find install: %w", err))
		return
	}

	s.createGrant(ctx, org.ID, app.GrantResourceTypeInstall, inst.ID)
}

// ListInstallGrants godoc
//
//	@ID				ListInstallGrants
//	@Summary		list grants on an install
//	@Tags			grants
//	@Produce		json
//	@Security		APIKey
//	@Security		OrgID
//	@Param			install_id	path		string	true	"install ID"
//	@Success		200			{array}		app.ResourceGrant
//	@Router			/v1/installs/{install_id}/grants [GET]
func (s *service) ListInstallGrants(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	inst := app.Install{}
	if err := s.resolveByNameOrID(ctx, org.ID, ctx.Param("install_id"), &inst); err != nil {
		ctx.Error(fmt.Errorf("unable to find install: %w", err))
		return
	}

	s.listGrants(ctx, org.ID, app.GrantResourceTypeInstall, inst.ID)
}

// DeleteInstallGrant godoc
//
//	@ID				DeleteInstallGrant
//	@Summary		revoke a grant on an install
//	@Tags			grants
//	@Security		APIKey
//	@Security		OrgID
//	@Param			install_id	path	string	true	"install ID"
//	@Param			grant_id	path	string	true	"grant ID"
//	@Success		204
//	@Router			/v1/installs/{install_id}/grants/{grant_id} [DELETE]
func (s *service) DeleteInstallGrant(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.requireOrgAdmin(ctx, org.ID); err != nil {
		ctx.Error(err)
		return
	}

	s.deleteGrant(ctx, org.ID, ctx.Param("grant_id"))
}

// CreateAppGrant godoc
//
//	@ID				CreateAppGrant
//	@Summary		grant an account access to an app
//	@Description	Grant an account read or full access to a single app (and its installs via walk-up). Org-admin only.
//	@Tags			grants
//	@Accept			json
//	@Produce		json
//	@Security		APIKey
//	@Security		OrgID
//	@Param			app_id	path		string				true	"app ID"
//	@Param			req		body		CreateGrantRequest	true	"grant"
//	@Success		201		{object}	app.ResourceGrant
//	@Router			/v1/apps/{app_id}/grants [POST]
func (s *service) CreateAppGrant(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.requireOrgAdmin(ctx, org.ID); err != nil {
		ctx.Error(err)
		return
	}

	a := app.App{}
	if err := s.resolveByNameOrID(ctx, org.ID, ctx.Param("app_id"), &a); err != nil {
		ctx.Error(fmt.Errorf("unable to find app: %w", err))
		return
	}

	s.createGrant(ctx, org.ID, app.GrantResourceTypeApp, a.ID)
}

// ListAppGrants godoc
//
//	@ID				ListAppGrants
//	@Summary		list grants on an app
//	@Tags			grants
//	@Produce		json
//	@Security		APIKey
//	@Security		OrgID
//	@Param			app_id	path		string	true	"app ID"
//	@Success		200		{array}		app.ResourceGrant
//	@Router			/v1/apps/{app_id}/grants [GET]
func (s *service) ListAppGrants(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	a := app.App{}
	if err := s.resolveByNameOrID(ctx, org.ID, ctx.Param("app_id"), &a); err != nil {
		ctx.Error(fmt.Errorf("unable to find app: %w", err))
		return
	}

	s.listGrants(ctx, org.ID, app.GrantResourceTypeApp, a.ID)
}

// DeleteAppGrant godoc
//
//	@ID				DeleteAppGrant
//	@Summary		revoke a grant on an app
//	@Tags			grants
//	@Security		APIKey
//	@Security		OrgID
//	@Param			app_id		path	string	true	"app ID"
//	@Param			grant_id	path	string	true	"grant ID"
//	@Success		204
//	@Router			/v1/apps/{app_id}/grants/{grant_id} [DELETE]
func (s *service) DeleteAppGrant(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.requireOrgAdmin(ctx, org.ID); err != nil {
		ctx.Error(err)
		return
	}

	s.deleteGrant(ctx, org.ID, ctx.Param("grant_id"))
}

// requireOrgAdmin permits only callers with org-wide administrative access.
// An app- or install-scoped `all` holder does not qualify (§9: org-admin only).
func (s *service) requireOrgAdmin(ctx *gin.Context, orgID string) error {
	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		return stderr.ErrAuthorization{
			Err:         fmt.Errorf("no account identified"),
			Description: "no account was set in the middleware",
		}
	}
	if err := acct.AllPermissions.CanPerform(orgID, permissions.PermissionAll); err != nil {
		return stderr.ErrAuthorization{
			Err:         fmt.Errorf("account is not an org admin"),
			Description: "only org admins may manage grants",
		}
	}
	return nil
}

func (s *service) createGrant(ctx *gin.Context, orgID string, resourceType app.GrantResourceType, resourceID string) {
	var req CreateGrantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := s.v.Struct(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	acctID, err := s.resolveGranteeID(ctx, req)
	if err != nil {
		ctx.Error(err)
		return
	}

	grant := app.ResourceGrant{
		OrgID:        orgID,
		AccountID:    acctID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Permission:   req.Permission,
	}
	res := s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "org_id"}, {Name: "account_id"}, {Name: "resource_type"}, {Name: "resource_id"}, {Name: "deleted_at"}},
			DoUpdates: clause.AssignmentColumns([]string{"permission", "updated_at"}),
		}).
		Create(&grant)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to create grant: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusCreated, grant)
}

func (s *service) listGrants(ctx *gin.Context, orgID string, resourceType app.GrantResourceType, resourceID string) {
	var grants []app.ResourceGrant
	res := s.db.WithContext(ctx).
		Where(app.ResourceGrant{OrgID: orgID, ResourceType: resourceType, ResourceID: resourceID}).
		Find(&grants)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to list grants: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, grants)
}

func (s *service) deleteGrant(ctx *gin.Context, orgID, grantID string) {
	res := s.db.WithContext(ctx).
		Where(app.ResourceGrant{OrgID: orgID}).
		Delete(&app.ResourceGrant{ID: grantID})
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to delete grant: %w", res.Error))
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (s *service) resolveGranteeID(ctx *gin.Context, req CreateGrantRequest) (string, error) {
	if req.AccountID == "" && req.Email == "" {
		return "", fmt.Errorf("one of account_id or email is required")
	}

	var acct app.Account
	q := s.db.WithContext(ctx)
	if req.AccountID != "" {
		q = q.Where(app.Account{ID: req.AccountID})
	} else {
		q = q.Where(app.Account{Email: req.Email})
	}
	if err := q.First(&acct).Error; err != nil {
		return "", fmt.Errorf("unable to find grantee account: %w", err)
	}
	return acct.ID, nil
}

func (s *service) resolveByNameOrID(ctx *gin.Context, orgID, nameOrID string, dst interface{}) error {
	return s.db.WithContext(ctx).
		Where("org_id = ?", orgID).
		Where(s.db.Where("id = ?", nameOrID).Or("name = ?", nameOrID)).
		First(dst).Error
}
