package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateComponentRequest struct {
	Name string `json:"name" validate:"required"`
}

func (c *CreateComponentRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @BasePath /v1/apps
// Create an app component
// @Summary create an app component
// @Schemes
// @Description create an app component
// @Param app_id path string true "app ID"
// @Param req body CreateComponentRequest true "Input"
// @Tags components
// @Accept json
// @Produce json
// @Success 201 {object} app.Component
// @Router /v1/apps/{app_id}/components/ [post]
func (s *service) CreateComponent(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req CreateComponentRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	component, err := s.createComponent(ctx, appID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create component: %w", err))
		return
	}

	s.hooks.Created(ctx, component.ID)
	ctx.JSON(http.StatusCreated, component)
}

func (s *service) createComponent(ctx context.Context, appID string, req *CreateComponentRequest) (*app.Component, error) {
	component := app.Component{
		Name: req.Name,
	}
	parentApp := app.App{}
	err := s.db.WithContext(ctx).Preload("Installs").First(&parentApp, "id = ?", appID).Association("Components").Append(&component)
	if err != nil {
		return nil, fmt.Errorf("unable to create component: %w", err)
	}
	if len(parentApp.Installs) < 1 {
		return &component, nil
	}

	// create an install component for all known installs
	var installCmps = []app.InstallComponent{}
	for _, install := range parentApp.Installs {
		installCmps = append(installCmps, app.InstallComponent{
			ComponentID: component.ID,
			InstallID:   install.ID,
		})
	}
	res := s.db.WithContext(ctx).Create(&installCmps)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create install components: %w", res.Error)
	}

	return &component, nil
}
