package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type UpdateAppRequest struct {
	Name string `json:"name"`
}

func (c *UpdateAppRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID UpdateApp
// @Summary	update an app
// @Description.markdown	update_app.md
// @Param			app_id	path	string				true	"app ID"
// @Param			req		body	UpdateAppRequest	true	"Input"
// @Tags			apps
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{object}	app.App
// @Router			/v1/apps/{app_id} [patch]
func (s *service) UpdateApp(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req UpdateAppRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse update request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	app, err := s.updateApp(ctx, appID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get  app%s: %w", appID, err))
		return
	}

	ctx.JSON(http.StatusOK, app)
}

func (s *service) updateApp(ctx context.Context, appID string, req *UpdateAppRequest) (*app.App, error) {
	currentApp := app.App{
		ID: appID,
	}

	res := s.db.WithContext(ctx).Model(&currentApp).Updates(app.App{Name: req.Name})
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update app: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return nil, fmt.Errorf("app not found %s %w", appID, gorm.ErrRecordNotFound)
	}

	return &currentApp, nil
}
