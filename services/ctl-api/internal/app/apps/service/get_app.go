package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
)

// @ID GetApp
// @Summary	get an app
// @Description.markdown	get_app.md
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
// @Success		200				{object}	app.App
// @Router			/v1/apps/{app_id} [get]
func (s *service) GetApp(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	app, err := s.findApp(ctx, org.ID, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app %s: %w", appID, err))
		return
	}

	ctx.JSON(http.StatusOK, app)
}

func (s *service) findApp(ctx context.Context, orgID, appID string) (*app.App, error) {
	app := app.App{}
	res := s.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("Org").
		Preload("Components").
		Preload("AppConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_configs_view.created_at DESC")
		}).

		//
		Preload("AppInputConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_input_configs.created_at DESC")
		}).
		Preload("AppInputConfigs.AppInputs").
		Preload("AppInputConfigs.AppInputs.AppInputGroup").
		Preload("AppInputConfigs.AppInputGroups.AppInputs").

		// runner config
		Preload("AppRunnerConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_runner_configs.created_at DESC")
		}).

		// sandbox configs
		Preload("AppSandboxConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_sandbox_configs.created_at DESC")
		}).
		Preload("AppSandboxConfigs.PublicGitVCSConfig").
		Preload("AppSandboxConfigs.ConnectedGithubVCSConfig").
		Preload("AppSandboxConfigs.SandboxRelease").
		Preload("AppSandboxConfigs.SandboxRelease.Sandbox").
		Where("name = ? AND org_id = ?", appID, orgID).
		Or("id = ?", appID).
		First(&app)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return &app, nil
}
