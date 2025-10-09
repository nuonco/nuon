package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
)

// @ID						GetApp
// @Summary				get an app
// @Description.markdown	get_app.md
// @Param					app_id	path	string	true	"app ID"
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.App
// @Router					/v1/apps/{app_id} [get]
func (s *service) GetApp(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	a, err := s.findApp(ctx, org.ID, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app %s: %w", appID, err))
		return
	}

	ctx.JSON(http.StatusOK, a)
}

func (s *service) appByNameOrID(ctx context.Context, appID string) (*app.App, error) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var currentApp app.App
	res := s.db.WithContext(ctx).
		Where("name = ? AND org_id = ?", appID, org.ID).
		Or("id = ?", appID).
		First(&currentApp)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to find app")
	}

	return &currentApp, nil
}

func (s *service) findApp(ctx context.Context, orgID, appID string) (*app.App, error) {
	a := app.App{}
	res := s.db.WithContext(ctx).
		Preload("Org").
		Preload("Components").
		Preload("AppConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order(views.TableOrViewName(s.db, &app.AppConfig{}, ".created_at DESC")).Limit(3)
		}).
		Preload("AppInputConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_input_configs.created_at DESC").Limit(5)
		}).
		Preload("AppInputConfigs.AppInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_inputs.index ASC")
		}).
		Preload("AppInputConfigs.AppInputs.AppInputGroup", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_input_groups.index ASC")
		}).
		Preload("AppInputConfigs.AppInputGroups.AppInputs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_inputs.index ASC")
		}).

		// runner config
		Preload("AppRunnerConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_runner_configs.created_at DESC").Limit(5)
		}).

		// sandbox configs
		Preload("AppSandboxConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_sandbox_configs.created_at DESC").Limit(5)
		}).
		Preload("AppSandboxConfigs.PublicGitVCSConfig").
		Preload("AppSandboxConfigs.ConnectedGithubVCSConfig").
		Preload("NotificationsConfig").
		Where("name = ? AND org_id = ?", appID, orgID).
		Or("id = ?", appID).
		First(&a)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return &a, nil
}
