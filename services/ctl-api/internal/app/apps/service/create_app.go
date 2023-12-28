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

type CreateAppRequest struct {
	Name string `json:"name" validate:"required"`
}

func (c *CreateAppRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID CreateApp
// @Summary	create an app
// @Description.markdown	create_app.md
// @Tags			apps
// @Accept			json
// @Param			req	body	CreateAppRequest	true	"Input"
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		201				{object}	app.App
// @Router			/v1/apps [post]
func (s *service) CreateApp(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req CreateAppRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	app, err := s.createApp(ctx, org.ID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app: %w", err))
		return
	}

	s.hooks.Created(ctx, app.ID, org.SandboxMode)
	ctx.JSON(http.StatusCreated, app)
}

func (s *service) createApp(ctx context.Context, orgID string, req *CreateAppRequest) (*app.App, error) {
	app := app.App{
		OrgID:             orgID,
		Name:              req.Name,
		Status:            "queued",
		StatusDescription: "waiting for event loop to start and provision app",
	}

	res := s.db.WithContext(ctx).
		Create(&app)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create app: %w", res.Error)
	}

	return &app, nil
}
