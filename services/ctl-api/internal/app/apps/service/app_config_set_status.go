package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
)

type AppLatestConfigSetStatusRequest struct {
	Status            app.AppConfigStatus `json:"status"`
	StatusDescription string              `json:"status_description"`
}

func (c *AppLatestConfigSetStatusRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID AppLatestConfigSetStatus
// @Summary	updates an app config sync status
// @Description.markdown	app_config_set_status.md
// @Tags			apps
// @Accept			json
// @Param			req	body	AppLatestConfigSetStatusRequest	true	"Input"
// @Produce		json
// @Param			app_id	path	string	true	"app ID"
// @Param			app_config_id	path	string	true	"app config ID"
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{boolean}	true
// @Router			/v1/apps/{app_id}/config/{app_config_id}/set-status [POST]
func (s *service) AppLatestConfigSetStatus(ctx *gin.Context) {
	org, err := middlewares.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req AppLatestConfigSetStatusRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	appID := ctx.Param("app_id")
	appConfigID := ctx.Param("app_config_id")

	err = s.updateStatus(ctx, appID, appConfigID, org, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to set app config status: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) updateStatus(ctx context.Context, appID string, appConfigID string, org *app.Org, req *AppLatestConfigSetStatusRequest) error {
	res := s.db.WithContext(ctx).
		Model(&app.AppConfig{}).
		Where(&app.AppConfig{
			ID:    appConfigID,
			AppID: appID,
			OrgID: org.ID,
		}).
		Updates(app.AppConfig{
			Status:            req.Status,
			StatusDescription: req.StatusDescription,
		})
	if res.Error != nil {
		return fmt.Errorf("unable to set app config status: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("app not found %s %w", appID, gorm.ErrRecordNotFound)
	}

	return nil
}
