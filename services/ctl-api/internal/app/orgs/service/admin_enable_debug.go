package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type DebugModeRequest struct {
	DebugMode bool `json:"debug"`
}

// @ID AdminDebugOrg
// @Summary control debug mode an org
// @Description.markdown debug_mode_org.md
// @Param			org_id	path	string	true	"org ID for your current org"
// @Tags			orgs/admin
// @Security AdminEmail
// @Accept			json
// @Param			req	body	DebugModeRequest	true	"Input"
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/orgs/{org_id}/admin-debug-mode [POST]
func (s *service) AdminDebugModeOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	org, err := s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req DebugModeRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	if err := s.updateDebugModeOrg(ctx, org.ID, req.DebugMode); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) updateDebugModeOrg(ctx context.Context, orgID string, debug bool) error {
	org := app.Org{
		ID: orgID,
	}
	res := s.db.WithContext(ctx).
		Model(&org).
		Updates(map[string]any{
			"debug_mode": debug,
		})
	if res.Error != nil {
		return fmt.Errorf("unable to update org: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("org not found %w", gorm.ErrRecordNotFound)
	}

	return nil
}
