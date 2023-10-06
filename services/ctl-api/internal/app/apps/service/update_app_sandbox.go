package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateAppSandboxRequest struct {
	SandboxReleaseID string `json:"sandbox_release_id" validate:"required"`
}

func (c *UpdateAppSandboxRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

//	@BasePath	/v1/apps
//
// Update app sandbox release
//
//	@Summary	update an app sandbox release
//	@Schemes
//	@Description	update an app sandbox release
//	@Param			app_id	path	string					true	"app ID"
//	@Param			req		body	UpdateAppSandboxRequest	true	"Input"
//	@Tags			apps
//	@Accept			json
//	@Produce		json
//	@Param			X-Nuon-Org-ID	header		string	true	"org ID"
//	@Param			Authorization	header		string	true	"bearer auth token"
//	@Failure		400				{object}	stderr.ErrResponse
//	@Failure		401				{object}	stderr.ErrResponse
//	@Failure		403				{object}	stderr.ErrResponse
//	@Failure		404				{object}	stderr.ErrResponse
//	@Failure		500				{object}	stderr.ErrResponse
//	@Success		200				{object}	app.App
//	@Router			/v1/apps/{app_id}/sandbox [PUT]
func (s *service) UpdateAppSandbox(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req UpdateAppSandboxRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse update request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	app, err := s.updateAppSandbox(ctx, appID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update app %s: %w", appID, err))
		return
	}

	s.hooks.SandboxReleaseUpdated(ctx, appID)
	ctx.JSON(http.StatusOK, app)
}

func (s *service) updateAppSandbox(ctx context.Context, appID string, req *UpdateAppSandboxRequest) (*app.App, error) {
	currentApp := app.App{
		ID: appID,
	}

	res := s.db.WithContext(ctx).Preload("SandboxRelease").Model(&currentApp).Updates(app.App{SandboxReleaseID: req.SandboxReleaseID})
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return &currentApp, nil
}
