package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type CreateComponentRequest struct {
	Name         string   `json:"name" validate:"required,interpolated_name"`
	VarName      string   `json:"var_name" validate:"interpolated_name"`
	Dependencies []string `json:"dependencies"`
}

func (c *CreateComponentRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						CreateComponent
// @Summary				create a component
// @Description.markdown	create_component.md
// @Param					app_id	path	string					true	"app ID"
// @Param					req		body	CreateComponentRequest	true	"Input"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.Component
// @Router					/v1/apps/{app_id}/components [post]
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

	s.evClient.Send(ctx, component.ID, &signals.Signal{
		Type: signals.OperationCreated,
	})
	s.evClient.Send(ctx, component.ID, &signals.Signal{
		Type: signals.OperationProvision,
	})
	s.evClient.Send(ctx, component.ID, &signals.Signal{
		Type: signals.OperationPollDependencies,
	})
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

	depIDs, err := s.helpers.GetComponentIDs(ctx, appID, req.Dependencies)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get component ids")
	}
	if err := s.helpers.CreateComponentDependencies(ctx, component.ID, depIDs); err != nil {
		return nil, fmt.Errorf("unable to create component dependencies: %w", err)
	}

	if err := s.helpers.EnsureInstallComponents(ctx, appID, nil); err != nil {
		return nil, fmt.Errorf("unable to ensure install components: %w", err)
	}

	return &component, nil
}
