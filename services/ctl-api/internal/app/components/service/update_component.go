package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateComponentRequest struct {
	Name string `json:"name"`
}

// @BasePath /v1/components
// Update a component
// @Summary update a component
// @Schemes
// @Description update a component
// @Param component_id path string component_id "component ID"
// @Param req body UpdateComponentRequest true "Input"
// @Tags components
// @Accept json
// @Produce json
// @Success 201 {object} app.Component
// @Router /v1/{component_id} [PATCH]
func (s *service) UpdateComponent(ctx *gin.Context) {
	var req UpdateComponentRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse update request: %w", err))
		return
	}

	componentID := ctx.Param("component_id")
	if componentID == "" {
		ctx.Error(fmt.Errorf("component_id must be passed in"))
		return
	}

	component, err := s.updateComponent(ctx, componentID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get  app%s: %w", componentID, err))
		return
	}

	ctx.JSON(http.StatusOK, component)
}

func (s *service) updateComponent(ctx context.Context, componentID string, req *UpdateComponentRequest) (*app.Component, error) {
	currentComponent := app.Component{
		ID: componentID,
	}

	res := s.db.WithContext(ctx).Model(&currentComponent).Updates(app.Component{Name: req.Name})
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component: %w", res.Error)
	}

	return &currentComponent, nil
}
