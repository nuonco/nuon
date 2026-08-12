package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						ListRoles
// @Summary				List your org's roles
// @Description.markdown	list_roles.md
// @Param					context	query	string	false	"filter to roles assignable on a surface (team, service_account, api_token, oidc_trust_policy)"	extensions(x-go-name=RoleContext)
// @Tags					roles
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
		Preload("Policies").
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
