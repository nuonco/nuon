package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @BasePath /v1/components
// Get all components
// @Summary get all components for all orgs
// @Schemes
// @Description get all components
// @Tags components/internal
// @Accept json
// @Produce json
// @Success 200 {array} app.Component
// @Router /v1/components [get]
func (s *service) GetAllComponents(ctx *gin.Context) {
	components, err := s.getAllComponents(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get components for: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, components)
}

func (s *service) getAllComponents(ctx context.Context) ([]*app.Component, error) {
	var components []*app.Component
	res := s.db.WithContext(ctx).Find(&components)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get all components: %w", res.Error)
	}

	return components, nil
}
