package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
)

type CreateComponentRequest struct {
	Name         string   `json:"name" validate:"required,interpolatedName"`
	VarName      string   `json:"var_name" validate:"interpolatedName"`
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

	// create component
	component, err := s.createComponent(ctx, appID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create component: %w", err))
		return
	}

	// validate to make sure graph does not have cycles
	if err = s.appsHelpers.ValidateGraph(ctx, appID); err != nil {
		ctx.Error(fmt.Errorf("invalid graph: %w", err))
		return
	}

	s.hooks.Created(ctx, component.ID, org.OrgType)
	ctx.JSON(http.StatusCreated, component)
}

func (s *service) createComponent(ctx context.Context, appID string, req *CreateComponentRequest) (*app.Component, error) {
	component := app.Component{
		AppID:             appID,
		Name:              req.Name,
		VarName:           req.VarName,
		Status:            "queued",
		StatusDescription: "waiting for event loop to start for component",
	}
	res := s.db.WithContext(ctx).
		Create(&component)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create component: %w", res.Error)
	}

	if err := s.helpers.CreateComponentDependencies(ctx, component.ID, req.Dependencies); err != nil {
		return nil, fmt.Errorf("unable to create component dependencies: %w", err)
	}

	if err := s.helpers.EnsureInstallComponents(ctx, appID, nil); err != nil {
		return nil, fmt.Errorf("unable to ensure install components: %w", err)
	}

	return &component, nil
}
