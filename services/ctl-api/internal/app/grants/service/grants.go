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

// CreateGrantRequest grants an account a permission on a single resource in the
// caller's org. The resource is any tier of the org -> app -> install spine, or
// a delegable org-owned resource (webhook, vcs_connection, slack_subscription);
// an org grant confers the permission on every resource in the org via walk-up.
// Exactly one of AccountID or Email identifies the grantee.
type CreateGrantRequest struct {
	ResourceType string `json:"resource_type" validate:"required,oneof=org app install webhook vcs_connection slack_subscription"`
	ResourceID   string `json:"resource_id" validate:"required"`
	AccountID    string `json:"account_id"`
	Email        string `json:"email"`
	Permission   string `json:"permission" validate:"required,oneof=read all"`
}

// CreateGrant godoc
//
//	@ID				CreateGrant
//	@Summary		grant an account access to a resource
//	@Description	Grant an account read or full access to a single resource (org, app, install, webhook, vcs_connection, or slack_subscription). An org grant covers every resource in the org, and an app grant covers its installs, via walk-up authorization. A resource_id of "*" covers every resource of that type in the org. Org-admin only.
//	@Tags			grants
//	@Accept			json
//	@Produce		json
//	@Security		APIKey
//	@Security		OrgID
//	@Param			req	body		CreateGrantRequest	true	"grant"
//	@Success		201	{object}	app.ResourceGrant
//	@Failure		400	{object}	stderr.ErrResponse
//	@Failure		403	{object}	stderr.ErrResponse
//	@Router			/v1/grants [POST]
func (s *service) CreateGrant(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.requireOrgAdmin(ctx, org.ID); err != nil {
		ctx.Error(err)
		return
	}

	var req CreateGrantRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := s.v.Struct(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	resourceType := app.GrantResourceType(req.ResourceType)
	resourceID, err := s.resolveResource(ctx, org.ID, resourceType, req.ResourceID)
	if err != nil {
		ctx.Error(err)
		return
	}

	acctID, err := s.resolveGranteeID(ctx, req)
	if err != nil {
		ctx.Error(err)
		return
	}

	grant := app.ResourceGrant{
		OrgID:        org.ID,
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

// ListGrants godoc
//
//	@ID				ListGrants
//	@Summary		list grants in an org
//	@Description	List grants in the caller's org, optionally filtered to a single resource by resource_type and resource_id.
//	@Tags			grants
//	@Produce		json
//	@Security		APIKey
//	@Security		OrgID
//	@Param			resource_type	query		string	false	"filter by resource type (org, app, install)"
//	@Param			resource_id		query		string	false	"filter by resource ID"
//	@Success		200				{array}		app.ResourceGrant
//	@Router			/v1/grants [GET]
func (s *service) ListGrants(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	where := app.ResourceGrant{OrgID: org.ID}
	if rt := ctx.Query("resource_type"); rt != "" {
		where.ResourceType = app.GrantResourceType(rt)
	}
	if rid := ctx.Query("resource_id"); rid != "" {
		where.ResourceID = rid
	}

	var grants []app.ResourceGrant
	res := s.db.WithContext(ctx).Where(where).Find(&grants)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to list grants: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, grants)
}

// DeleteGrant godoc
//
//	@ID				DeleteGrant
//	@Summary		revoke a grant
//	@Tags			grants
//	@Security		APIKey
//	@Security		OrgID
//	@Param			grant_id	path	string	true	"grant ID"
//	@Success		204
//	@Router			/v1/grants/{grant_id} [DELETE]
func (s *service) DeleteGrant(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.requireOrgAdmin(ctx, org.ID); err != nil {
		ctx.Error(err)
		return
	}

	res := s.db.WithContext(ctx).
		Where(app.ResourceGrant{OrgID: org.ID}).
		Delete(&app.ResourceGrant{ID: ctx.Param("grant_id")})
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to delete grant: %w", res.Error))
		return
	}

	ctx.Status(http.StatusNoContent)
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

// resolveResource validates that the named resource exists in the org and
// returns its canonical ID. Apps and installs are accepted by name or ID,
// org-owned resources (webhooks, VCS connections, Slack subscriptions) by ID
// only, and every non-org type accepts the "*" wildcard to scope the grant to
// every resource of that type in the org. The org target must be the caller's
// own org (no wildcard — the org is already a single resource).
func (s *service) resolveResource(ctx *gin.Context, orgID string, resourceType app.GrantResourceType, nameOrID string) (string, error) {
	switch resourceType {
	case app.GrantResourceTypeOrg:
		if nameOrID != orgID {
			return "", stderr.NewInvalidRequest(fmt.Errorf("org grant resource_id %q must be your own org %q", nameOrID, orgID))
		}
		return orgID, nil
	case app.GrantResourceTypeApp:
		if nameOrID == app.GrantResourceWildcard {
			return app.GrantResourceWildcard, nil
		}
		var a app.App
		if err := s.resolveByNameOrID(ctx, orgID, nameOrID, &a); err != nil {
			return "", fmt.Errorf("unable to find app: %w", err)
		}
		return a.ID, nil
	case app.GrantResourceTypeInstall:
		if nameOrID == app.GrantResourceWildcard {
			return app.GrantResourceWildcard, nil
		}
		var inst app.Install
		if err := s.resolveByNameOrID(ctx, orgID, nameOrID, &inst); err != nil {
			return "", fmt.Errorf("unable to find install: %w", err)
		}
		return inst.ID, nil
	case app.GrantResourceTypeWebhook:
		return s.resolveOrgResourceID(ctx, orgID, nameOrID, &app.Webhook{}, "webhook")
	case app.GrantResourceTypeVCSConnection:
		return s.resolveOrgResourceID(ctx, orgID, nameOrID, &app.VCSConnection{}, "vcs connection")
	case app.GrantResourceTypeSlackSubscription:
		return s.resolveOrgResourceID(ctx, orgID, nameOrID, &app.SlackChannelSubscription{}, "slack subscription")
	default:
		return "", stderr.NewInvalidRequest(fmt.Errorf("unsupported resource type %q; must be one of org, app, install, webhook, vcs_connection, slack_subscription", resourceType))
	}
}

func (s *service) resolveOrgResourceID(ctx *gin.Context, orgID, id string, dst interface{}, kind string) (string, error) {
	if id == app.GrantResourceWildcard {
		return app.GrantResourceWildcard, nil
	}
	err := s.db.WithContext(ctx).
		Where("org_id = ?", orgID).
		Where("id = ?", id).
		First(dst).Error
	if err != nil {
		return "", fmt.Errorf("unable to find %s: %w", kind, err)
	}
	return id, nil
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
