package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

const (
	roleAppliesToUser           = "user"
	roleAppliesToServiceAccount = "service_account"
)

type RoleInfo struct {
	RoleType    app.RoleType `json:"role_type"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	AppliesTo   []string     `json:"applies_to"`
}

var roleCatalog = []RoleInfo{
	{
		RoleType:    app.RoleTypeOrgAdmin,
		Title:       "Admin",
		Description: "Full access to the organization and all of its resources.",
		AppliesTo:   []string{roleAppliesToUser, roleAppliesToServiceAccount},
	},
	{
		RoleType:    app.RoleTypeOrgReadOnly,
		Title:       "Read-only",
		Description: "Read-only access to the organization and its resources.",
		AppliesTo:   []string{roleAppliesToUser, roleAppliesToServiceAccount},
	},
	{
		RoleType:    app.RoleTypeRunner,
		Title:       "Runner",
		Description: "Permissions for runners executing deployments.",
		AppliesTo:   []string{roleAppliesToServiceAccount},
	},
}

func roleAppliesTo(roleType app.RoleType, appliesTo string) bool {
	for _, info := range roleCatalog {
		if info.RoleType != roleType {
			continue
		}
		for _, a := range info.AppliesTo {
			if a == appliesTo {
				return true
			}
		}
	}
	return false
}

func validateServiceAccountRole(role string) (app.RoleType, error) {
	roleType := app.RoleType(role)
	if !roleAppliesTo(roleType, roleAppliesToServiceAccount) {
		return "", stderr.ErrUser{
			Err:         fmt.Errorf("invalid role %q for service account", role),
			Description: "role must be a valid service account role; see GET /v1/roles",
		}
	}

	return roleType, nil
}

// @ID						ListRoles
// @Summary				List assignable roles
// @Description.markdown	list_roles.md
// @Tags					accounts
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	[]RoleInfo
// @Router					/v1/roles [GET]
func (s *service) ListRoles(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, roleCatalog)
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

// @ID						ListServiceAccounts
// @Summary				List service accounts for the current org
// @Description.markdown	list_service_accounts.md
// @Param					offset			query	int		false	"offset of results to return"	Default(0)
// @Param					limit			query	int		false	"limit of results to return"	Default(10)
// @Param					page			query	int		false	"page number of results to return"	Default(0)
// @Param					include_runners	query	bool	false	"include service accounts with the runner role (excluded by default)"
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

	accounts := []app.Account{}
	tx := s.db.WithContext(ctx).
		Model(&app.Account{}).
		Joins("JOIN account_roles ON account_roles.account_id = accounts.id AND account_roles.org_id = ? AND account_roles.deleted_at = 0", org.ID).
		Where("accounts.account_type = ?", app.AccountTypeService)

	if !includeRunners {
		tx = tx.
			Joins("JOIN roles ON roles.id = account_roles.role_id AND roles.deleted_at = 0").
			Where("roles.role_type != ?", app.RoleTypeRunner)
	}

	tx = tx.
		Group("accounts.id").
		Order("accounts.email").
		Order("accounts.id").
		Scopes(scopes.WithOffsetPagination).
		Preload("Roles", "org_id = ?", org.ID).
		Preload("Roles.Org").
		Preload("Roles.Policies").
		Find(&accounts)
	if tx.Error != nil {
		ctx.Error(fmt.Errorf("unable to list service accounts for org %s: %w", org.ID, tx.Error))
		return
	}

	accounts, err = db.HandlePaginatedResponse(ctx, accounts)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to handle paginated response: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, accounts)
}

type CreateServiceAccountRequest struct {
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

	roleType, err := validateServiceAccountRole(req.Role)
	if err != nil {
		ctx.Error(err)
		return
	}

	acct, err := s.acctClient.CreateServiceAccount(ctx, domains.NewAccountID())
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

	roleType, err := validateServiceAccountRole(req.Role)
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

	token, err := s.acctClient.CreateToken(ctx, acct.Email, dur)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create token: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, CreateServiceAccountTokenResponse{
		Token: token.Token,
	})
}
