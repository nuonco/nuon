package service

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateRoleRequest struct {
	// Title is the role's display name.
	Title string `json:"title" binding:"required"`
	// Description explains what the role grants.
	Description string `json:"description"`
	// Contexts names the assignment surfaces the role may be offered on
	// (team, service_account, api_token, oidc_trust_policy).
	Contexts []string `json:"contexts"`
	// Permissions are the scoped permission entries the role's policy carries.
	Permissions []PermissionEntryRequest `json:"permissions" binding:"required"`
}

// @ID						CreateRole
// @Summary				Create a custom role for the current org
// @Description			Create a role whose policy carries scoped permission entries. Custom roles are assigned through the same flows as managed roles, addressed by role id.
// @Param					req	body	CreateRoleRequest	true	"Input"
// @Tags					roles
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.Role
// @Router					/v1/roles [POST]
func (s *service) CreateRole(ctx *gin.Context) {
	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req CreateRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(userErr(fmt.Errorf("unable to parse request: %w", err)))
		return
	}

	if err := validateContexts(req.Contexts); err != nil {
		ctx.Error(userErr(err))
		return
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		ctx.Error(userErr(fmt.Errorf("title is required")))
		return
	}
	if err := s.requireAvailableTitle(ctx, org.ID, title, ""); err != nil {
		ctx.Error(err)
		return
	}

	entries, err := s.validatePermissionEntries(ctx, org, req.Permissions)
	if err != nil {
		ctx.Error(err)
		return
	}

	role := app.Role{
		OrgID:       generics.NewNullString(org.ID),
		RoleType:    app.RoleTypeCustom,
		Title:       title,
		Description: req.Description,
		Contexts:    req.Contexts,
		Managed:     false,
		Policies: []app.Policy{
			{
				OrgID:             generics.NewNullString(org.ID),
				Name:              app.PolicyNameCustom,
				ScopedPermissions: entries,
			},
		},
	}

	if res := s.db.WithContext(ctx).Create(&role); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to create role: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusCreated, role)
}

func validateContexts(contexts []string) error {
	known := []string{
		app.RoleContextTeam,
		app.RoleContextServiceAccount,
		app.RoleContextAPIToken,
		app.RoleContextTrustPolicy,
	}
	for _, c := range contexts {
		if !slices.Contains(known, c) {
			return fmt.Errorf("invalid context %q: must be one of %v", c, known)
		}
	}
	return nil
}
