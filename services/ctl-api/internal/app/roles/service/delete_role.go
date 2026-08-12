package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @ID						DeleteRole
// @Summary				Delete a custom role for the current org
// @Description			Delete a custom role, revoking it from every account it is assigned to. Managed roles cannot be deleted.
// @Param					role_id	path	string	true	"role ID"
// @Tags					roles
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				204
// @Router					/v1/roles/{role_id} [DELETE]
func (s *service) DeleteRole(ctx *gin.Context) {
	org, err := s.requireOrgAdmin(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	role, err := s.requireCustomRole(ctx, org.ID, ctx.Param("role_id"))
	if err != nil {
		ctx.Error(err)
		return
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().
			Where(app.AccountRole{RoleID: role.ID}).
			Delete(&app.AccountRole{}).Error; err != nil {
			return fmt.Errorf("unable to revoke role assignments: %w", err)
		}

		if err := tx.
			Where(app.Policy{RoleID: role.ID}).
			Delete(&app.Policy{}).Error; err != nil {
			return fmt.Errorf("unable to delete role policies: %w", err)
		}

		if err := tx.Delete(&app.Role{ID: role.ID}).Error; err != nil {
			return fmt.Errorf("unable to delete role: %w", err)
		}

		return nil
	})
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}
