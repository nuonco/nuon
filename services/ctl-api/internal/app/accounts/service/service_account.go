package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

func (s *service) resolveServiceAccountRole(ctx *gin.Context, orgID, role string) (app.RoleType, error) {
	resolved, err := s.authzClient.ResolveAssignableRole(ctx, orgID, app.RoleType(role), app.RoleContextServiceAccount)
	if err != nil {
		return "", stderr.ErrUser{
			Err:         err,
			Description: err.Error(),
		}
	}
	return resolved.RoleType, nil
}

// @ID						ListRoles
// @Summary				List your org's roles
// @Description.markdown	list_roles.md
// @Param					context	query	string	false	"filter to roles assignable on a surface (team, service_account, api_token, oidc_trust_policy)"	extensions(x-go-name=RoleContext)
// @Tags					accounts
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	[]app.Role
// @Router					/v1/roles [GET]
func (s *service) ListRoles(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var roles []app.Role
	res := s.db.WithContext(ctx).
		Where(app.Role{OrgID: generics.NewNullString(org.ID)}).
		Order("role_type").
		Find(&roles)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to list roles for org %s: %w", org.ID, res.Error))
		return
	}

	if roleContext := ctx.Query("context"); roleContext != "" {
		filtered := make([]app.Role, 0, len(roles))
		for _, role := range roles {
			if role.AllowsContext(roleContext) {
				filtered = append(filtered, role)
			}
		}
		roles = filtered
	}

	ctx.JSON(http.StatusOK, roles)
}

// getOrgServiceAccount looks up an account by ID and ensures it is a service
// account that belongs to the given org.
func (s *service) getOrgServiceAccount(ctx context.Context, orgID, accountID string) (*app.Account, error) {
	acct, err := s.acctClient.FindAccount(ctx, accountID)
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

// orgServiceAccountIDs resolves the org's service accounts off the
// account_roles org index. Joining accounts to account_roles in the list query
// instead lets the planner walk the whole accounts table in email order to
// satisfy the ORDER BY and LIMIT, which took tens of seconds on a cold cache.
func (s *service) orgServiceAccountIDs(ctx context.Context, orgID string, includeRunners, includeStacks bool) ([]string, error) {
	tx := s.db.WithContext(ctx).
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

	accountIDs := []string{}
	if err := tx.Distinct().Pluck("account_roles.account_id", &accountIDs).Error; err != nil {
		return nil, fmt.Errorf("unable to list service account ids for org %s: %w", orgID, err)
	}

	return accountIDs, nil
}

// @ID						ListServiceAccounts
// @Summary				List service accounts for the current org
// @Description.markdown	list_service_accounts.md
// @Param					offset			query	int		false	"offset of results to return"	Default(0)
// @Param					limit			query	int		false	"limit of results to return"	Default(10)
// @Param					page			query	int		false	"page number of results to return"	Default(0)
// @Param					include_runners	query	bool	false	"include service accounts with the runner role (excluded by default)"
// @Param					include_stacks	query	bool	false	"include service accounts with the stack role (excluded by default)"
// @Tags					accounts
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	[]app.Account
// @Router					/v1/service-accounts [GET]
func (s *service) ListServiceAccounts(ctx *gin.Context) {
	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	includeRunners := ctx.Query("include_runners") == "true"
	includeStacks := ctx.Query("include_stacks") == "true"

	accountIDs, err := s.orgServiceAccountIDs(ctx, org.ID, includeRunners, includeStacks)
	if err != nil {
		ctx.Error(err)
		return
	}

	accounts := []app.Account{}
	if len(accountIDs) > 0 {
		res := s.db.WithContext(ctx).
			Model(&app.Account{}).
			Where("accounts.id IN ?", accountIDs).
			Order("accounts.email").
			Order("accounts.id").
			Scopes(scopes.WithOffsetPagination).
			Preload("Roles", "org_id = ?", org.ID).
			Preload("Roles.Org").
			Preload("Roles.Policies").
			Find(&accounts)
		if res.Error != nil {
			ctx.Error(fmt.Errorf("unable to list service accounts for org %s: %w", org.ID, res.Error))
			return
		}
	}

	accounts, err = db.HandlePaginatedResponse(ctx, accounts)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to handle paginated response: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, accounts)
}

type CreateServiceAccountRequest struct {
	// Name is a human-friendly label for the service account.
	Name string `json:"name" binding:"required"`
	// Role must be one of the service account roles returned by GET /v1/roles.
	Role string `json:"role" binding:"required"`
}

// @ID						CreateServiceAccount
// @Summary				Create a service account for the current org
// @Description.markdown	create_service_account.md
// @Param					req	body	CreateServiceAccountRequest	true	"Input"
// @Tags					accounts
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.Account
// @Router					/v1/service-accounts [POST]
func (s *service) CreateServiceAccount(ctx *gin.Context) {
	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req CreateServiceAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse request: %w", err),
			Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
		})
		return
	}

	roleType, err := s.resolveServiceAccountRole(ctx, org.ID, req.Role)
	if err != nil {
		ctx.Error(err)
		return
	}

	acct, err := s.acctClient.CreateServiceAccount(ctx, domains.NewAccountID(), req.Name)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create service account: %w", err))
		return
	}

	if err := s.authzClient.SetAccountOrgRole(ctx, org.ID, acct.ID, roleType); err != nil {
		ctx.Error(fmt.Errorf("unable to assign role to service account: %w", err))
		return
	}

	acct, err = s.acctClient.FindAccount(ctx, acct.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to reload service account: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, acct)
}

type UpdateServiceAccountRequest struct {
	// Name is a human-friendly label for the service account.
	Name string `json:"name" binding:"required"`
}

// @ID						UpdateServiceAccount
// @Summary				Update a service account for the current org
// @Description.markdown	update_service_account.md
// @Param					account_id	path	string						true	"service account ID"
// @Param					req			body	UpdateServiceAccountRequest	true	"Input"
// @Tags					accounts
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.Account
// @Router					/v1/service-accounts/{account_id} [PATCH]
func (s *service) UpdateServiceAccount(ctx *gin.Context) {
	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	accountID := ctx.Param("account_id")

	var req UpdateServiceAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse request: %w", err),
			Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
		})
		return
	}

	acct, err := s.getOrgServiceAccount(ctx, org.ID, accountID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.db.WithContext(ctx).
		Model(&app.Account{}).
		Where(app.Account{ID: acct.ID}).
		Update("name", req.Name).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to update service account: %w", err))
		return
	}

	acct, err = s.acctClient.FindAccount(ctx, acct.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to reload service account: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, acct)
}

type UpdateServiceAccountRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

// @ID						UpdateServiceAccountRole
// @Summary				Update the role of a service account for the current org
// @Description.markdown	update_service_account_role.md
// @Param					account_id	path	string							true	"service account ID"
// @Param					req			body	UpdateServiceAccountRoleRequest	true	"Input"
// @Tags					accounts
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.Account
// @Router					/v1/service-accounts/{account_id}/role [PATCH]
func (s *service) UpdateServiceAccountRole(ctx *gin.Context) {
	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	accountID := ctx.Param("account_id")

	var req UpdateServiceAccountRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse request: %w", err),
			Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
		})
		return
	}

	roleType, err := s.resolveServiceAccountRole(ctx, org.ID, req.Role)
	if err != nil {
		ctx.Error(err)
		return
	}

	acct, err := s.getOrgServiceAccount(ctx, org.ID, accountID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.authzClient.SetAccountOrgRole(ctx, org.ID, acct.ID, roleType); err != nil {
		ctx.Error(fmt.Errorf("unable to update service account role: %w", err))
		return
	}

	acct, err = s.acctClient.FindAccount(ctx, acct.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to reload service account: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, acct)
}

// @ID						DeleteServiceAccount
// @Summary				Delete a service account for the current org
// @Description.markdown	delete_service_account.md
// @Param					account_id	path	string	true	"service account ID"
// @Tags					accounts
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				202
// @Router					/v1/service-accounts/{account_id} [DELETE]
func (s *service) DeleteServiceAccount(ctx *gin.Context) {
	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	accountID := ctx.Param("account_id")

	acct, err := s.getOrgServiceAccount(ctx, org.ID, accountID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.authzClient.RemoveAccountOrgRoles(ctx, org.ID, acct.ID); err != nil {
		ctx.Error(fmt.Errorf("unable to remove roles: %w", err))
		return
	}

	if err := s.acctClient.InvalidateTokens(ctx, acct.Email); err != nil {
		ctx.Error(fmt.Errorf("unable to invalidate tokens: %w", err))
		return
	}

	ctx.Status(http.StatusAccepted)
}

type CreateServiceAccountTokenRequest struct {
	// Duration defaults to one year.
	Duration string `json:"duration" default:"8760h"`

	// Name labels the token where it is listed; defaults to the account's identity.
	Name string `json:"name"`

	Invalidate bool `json:"invalidate"`
}

type CreateServiceAccountTokenResponse struct {
	Token string `json:"token,omitzero"`
}

// @ID						CreateServiceAccountToken
// @Summary				Create a token for a service account in the current org
// @Description.markdown	create_service_account_token.md
// @Param					account_id	path	string								true	"service account ID"
// @Param					req			body	CreateServiceAccountTokenRequest	true	"Input"
// @Tags					accounts
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	CreateServiceAccountTokenResponse
// @Router					/v1/service-accounts/{account_id}/tokens [POST]
func (s *service) CreateServiceAccountToken(ctx *gin.Context) {
	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	accountID := ctx.Param("account_id")

	var req CreateServiceAccountTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, http.ErrBodyNotAllowed) {
		if err.Error() != "EOF" {
			ctx.Error(stderr.ErrUser{
				Err:         fmt.Errorf("unable to parse request: %w", err),
				Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
			})
			return
		}
	}

	duration := req.Duration
	if duration == "" {
		duration = "8760h"
	}

	dur, err := time.ParseDuration(duration)
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid duration: %w", err),
			Description: fmt.Sprintf("invalid duration: %s", err.Error()),
		})
		return
	}

	acct, err := s.getOrgServiceAccount(ctx, org.ID, accountID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if req.Invalidate {
		if err := s.acctClient.InvalidateTokens(ctx, acct.Email); err != nil {
			ctx.Error(fmt.Errorf("unable to invalidate tokens: %w", err))
			return
		}
	}

	caller, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = acct.Email
	}

	// createStaticToken, not acctClient.CreateToken: only this one stamps the columns
	// ListStaticTokens filters on, so tokens made the other way cannot be revoked
	// from the org's API tokens page.
	token, err := s.createStaticToken(ctx, acct, org.ID, caller.ID, name, orgRoleType(acct, org.ID), dur)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create token: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, CreateServiceAccountTokenResponse{
		Token: token.Token,
	})
}

// orgRoleType reports the account's org role for display only.
func orgRoleType(acct *app.Account, orgID string) app.RoleType {
	for _, role := range acct.Roles {
		if role.OrgID.ValueString() == orgID {
			return role.RoleType
		}
	}

	return ""
}
