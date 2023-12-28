package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateComponentRequest struct {
	Name string `json:"name" validate:"required,interpolatedName"`
}

func (c *UpdateComponentRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID UpdateComponent
// @Summary	update a component
// @Description.markdown	update_component.md
// @Param			component_id	path	string					true	"component ID"
// @Param			req				body	UpdateComponentRequest	true	"Input"
// @Tags			components
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{object}	app.Component
// @Router			/v1/components/{component_id} [PATCH]
func (s *service) UpdateComponent(ctx *gin.Context) {
	componentID := ctx.Param("component_id")
	var req UpdateComponentRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse update request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
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
