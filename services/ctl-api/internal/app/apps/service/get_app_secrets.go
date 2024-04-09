package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

// @ID GetAppSecrets
// @Summary	get app secrets
// @Description.markdown	get_app_secrets.md
// @Param			app_id	path	string	true	"app ID"
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
// @Success		200				{array}	app.AppSecret
// @Router			/v1/apps/{app_id}/secrets [get]
func (s *service) GetAppSecrets(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	secrets, err := s.getAppSecrets(ctx, appID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, secrets)
}

func (s *service) getAppSecrets(ctx context.Context, appID string) ([]app.AppSecret, error) {
	var currentApp app.App

	res := s.db.WithContext(ctx).
		Preload("AppSecrets.CreatedBy").
		Preload("AppSecrets", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_secrets.created_at DESC")
		}).
		First(&currentApp, "id = ?", appID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app secrets: %w", res.Error)
	}

	return currentApp.AppSecrets, nil
}
