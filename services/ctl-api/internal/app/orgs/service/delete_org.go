package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
)

// @BasePath /v1/orgs

// Delete an org
// @Summary Delete an org
// @Schemes
// @Description create a new org
// @Tags orgs
// @Accept json
// @Produce json
// @Success 200 {boolean} ok
// @Router /v1/orgs/current [DELETE]
func (s *service) DeleteOrg(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	err = s.deleteOrg(ctx, org.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.hooks.Deleted(ctx, org.ID)
	ctx.JSON(http.StatusOK, true)
}

func (s *service) deleteOrg(ctx context.Context, orgID string) error {
	res := s.db.WithContext(ctx).Delete(&app.Org{
		ID: orgID,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to delete org: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("org not found")
	}

	return nil
}
