package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @BasePath /v1/orgs

// Delete an org
// @Summary Delete an org
// @Schemes
// @Description create a new org
// @Param org_id path string org_id "org ID for your current org"
// @Tags orgs
// @Accept json
// @Produce json
// @Success 201 {string} ok
// @Router /v1/orgs/{org_id} [DELETE]
func (s *service) DeleteOrg(ctx *gin.Context) {
	orgID := ctx.Param("id")

	err := s.deleteOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.hooks.Deleted(ctx, orgID)
	ctx.JSON(http.StatusAccepted, map[string]string{
		"status": "ok",
	})
}

func (s *service) deleteOrg(ctx context.Context, orgID string) error {
	res := s.db.WithContext(ctx).Delete(&app.Org{
		ID: orgID,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to delete org: %w", res.Error)
	}

	return nil
}
