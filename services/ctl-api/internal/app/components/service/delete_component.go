package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @BasePath /v1/components
// Delete an component
// @Summary delete an component
// @Schemes
// @Description delete an component
// @Param component_id path string true "component ID"
// @Tags components
// @Accept json
// @Produce json
// @Success 200 {boolean} true
// @Router /v1/components/{component_id} [DELETE]
func (s *service) DeleteComponent(ctx *gin.Context) {
	componentID := ctx.Param("component_id")

	err := s.deleteComponent(ctx, componentID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.hooks.Deleted(ctx, componentID)
	ctx.JSON(http.StatusOK, true)
}

func (s *service) deleteComponent(ctx context.Context, componentID string) error {
	res := s.db.WithContext(ctx).Delete(&app.Component{
		ID: componentID,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to delete component: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("component not found")
	}

	return nil
}
