package service

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
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
func (s *service) GetApp(ctx *gin.Context) (any, error) {
	return s.getAuthorizedApp(ctx.Request.Context())
}

func (s *service) appByNameOrID(ctx context.Context, appID string) (*app.App, error) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		return nil, err
	}

	var currentApp app.App
	res := s.db.WithContext(ctx).
		Where("org_id = ?", org.ID).
		Where(s.db.Where("name = ?", appID).Or("id = ?", appID)).
		First(&currentApp)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to find app")
	}

	return &currentApp, nil
}

func (s *service) findApp(ctx context.Context, orgID, appID string) (*app.App, error) {
	return s.loadAppDetails(ctx, orgID, s.db.Where("name = ?", appID).Or("id = ?", appID))
}

func (s *service) loadAppDetails(ctx context.Context, orgID string, filter *gorm.DB) (*app.App, error) {
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
		Where("org_id = ?", orgID).
		Where(filter).
		First(&a)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return &a, nil
}

var appDetailRead = require.ReadContract{
	Permission: permissions.AppRead,
	Targets:    []require.ReadTarget{{Name: "app_id", Kind: permissions.KindApp}},
}

func (s *service) getAuthorizedApp(ctx context.Context) (*app.App, error) {
	access, err := require.ReadAccessFromContext(ctx)
	if err != nil {
		return nil, err
	}
	target, err := access.Target("app_id")
	if err != nil {
		return nil, err
	}
	reader := *s
	reader.db = access.DB()
	result, err := reader.loadAppDetails(ctx, target.OrgID, reader.db.Where(app.App{ID: target.ID}))
	if err != nil {
		return nil, err
	}
	result.NotificationsConfig.SlackWebhookURL = ""
	return result, nil
}
