package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
	"gorm.io/gorm/clause"
)

type CreateComponentRequest struct {
	Name         string   `json:"name" validate:"required,interpolatedName"`
	Dependencies []string `json:"dependencies"`
}

func (c *CreateComponentRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID CreateComponent
// @Summary	create a component
// @Description.markdown	create_component.md
// @Param			app_id	path	string					true	"app ID"
// @Param			req		body	CreateComponentRequest	true	"Input"
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
// @Success		201				{object}	app.Component
// @Router			/v1/apps/{app_id}/components [post]
func (s *service) CreateComponent(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

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

	s.hooks.Created(ctx, component.ID, org.SandboxMode)
	ctx.JSON(http.StatusCreated, component)
}

func (s *service) createComponent(ctx context.Context, appID string, req *CreateComponentRequest) (*app.Component, error) {
	component := app.Component{
		AppID:             appID,
		Name:              req.Name,
		Status:            "queued",
		StatusDescription: "waiting for event loop to start for component",
	}

	res := s.db.WithContext(ctx).
		Create(&component)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create component: %w", res.Error)
	}

	// create dependencies
	deps := make([]*app.Component, 0, len(req.Dependencies))
	for _, depID := range req.Dependencies {
		deps = append(deps, &app.Component{
			ID: depID,
		})
	}
	if err := s.db.WithContext(ctx).Model(&component).
		Association("Dependencies").
		Replace(deps); err != nil {
		return nil, fmt.Errorf("unable to replace dependencies: %w", err)
	}

	// fetch the parent app's installs and ensure each gets the new component
	parentApp := app.App{}
	res = s.db.WithContext(ctx).
		Preload("Installs").
		First(&parentApp, "id = ?", appID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create component: %w", res.Error)
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
	res = s.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&installCmps)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create install components: %w", res.Error)
	}

	return &component, nil
}
