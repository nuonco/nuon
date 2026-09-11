package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const appConfigIntermediateMaxAge = 300

// @ID						GetAppConfigIntermediate
// @Summary				get an app config's intermediate config
// @Description			Returns the parsed intermediate config for a specific app config version, plus a map of config names to database ids.
// @Param					app_id		path	string	true	"app ID"
// @Param					config_id	path	string	true	"app config ID"
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
// @Success				200	{object}	app.AppConfigIntermediate
// @Router					/v1/apps/{app_id}/configs/{config_id}/intermediate [get]
func (s *service) GetAppConfigIntermediate(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org: %w", err))
		return
	}

	appID := ctx.Param("app_id")
	configID := ctx.Param("config_id")

	appCfg, err := s.getAppConfig(ctx, org.ID, appID, configID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Header("Cache-Control", "private, max-age="+strconv.Itoa(appConfigIntermediateMaxAge))
	s.respondAppConfigIntermediate(ctx, appCfg)
}

// @ID						GetAppBranchIntermediateConfig
// @Summary				get an app branch's intermediate config
// @Description			Returns the parsed intermediate config for the branch's latest active config, plus a map of config names to database ids.
// @Param					app_id			path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
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
// @Success				200	{object}	app.AppConfigIntermediate
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/intermediate-config [get]
func (s *service) GetAppBranchIntermediateConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org: %w", err))
		return
	}

	appID := ctx.Param("app_id")
	branchID := ctx.Param("app_branch_id")

	var branch app.AppBranch
	if err := s.db.WithContext(ctx).
		Where(app.AppBranch{OrgID: org.ID, AppID: appID}).
		First(&branch, "id = ?", branchID).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to find app branch: %w", err))
		return
	}

	appCfg, err := s.helpers.GetLatestActiveAppConfigForBranch(ctx, appID, branchID)
	if err != nil {
		ctx.Error(err)
		return
	}

	// GetLatestActiveAppConfigForBranch silently falls back to the app's latest active config
	// when the branch has never synced.
	if appCfg.AppBranchID.String != branchID {
		ctx.Error(stderr.ErrNotFound{
			Err:         fmt.Errorf("no config has been synced for app branch %s: %w", branchID, gorm.ErrRecordNotFound),
			Description: "this branch has not synced a config yet",
		})
		return
	}

	if appCfg.OrgID != org.ID {
		ctx.Error(fmt.Errorf("app config not found: %w", gorm.ErrRecordNotFound))
		return
	}

	s.respondAppConfigIntermediate(ctx, appCfg)
}

func (s *service) respondAppConfigIntermediate(ctx *gin.Context, appCfg *app.AppConfig) {
	if appCfg.IntermediateConfig == nil || !appCfg.IntermediateConfig.IsSet() {
		ctx.Error(stderr.ErrNotFound{
			Err:         fmt.Errorf("app config %s has no intermediate config: %w", appCfg.ID, gorm.ErrRecordNotFound),
			Description: "this app config was created without an intermediate config",
		})
		return
	}

	blobCtx := blobstore.WithBlobService(ctx.Request.Context(), s.blobSvc)
	raw, err := appCfg.IntermediateConfig.Get(blobCtx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to load intermediate config: %w", err))
		return
	}

	var cfg config.AppConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		ctx.Error(fmt.Errorf("unable to parse intermediate config: %w", err))
		return
	}

	resources, err := s.getAppConfigResources(ctx, appCfg)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app config resources: %w", err))
		return
	}

	metadata := appCfg.IntermediateConfig.Metadata()
	if metadata.Checksum != "" {
		ctx.Header("ETag", strconv.Quote(metadata.Checksum))
	}

	ctx.JSON(http.StatusOK, app.AppConfigIntermediate{
		ConfigID:            appCfg.ID,
		AppID:               appCfg.AppID,
		AppBranchID:         appCfg.AppBranchID.String,
		Status:              appCfg.Status,
		StatusV2:            appCfg.StatusV2,
		Version:             appCfg.Version,
		CLIVersion:          appCfg.CLIVersion,
		Checksum:            appCfg.Checksum,
		Size:                metadata.Size,
		CreatedAt:           appCfg.CreatedAt,
		CreatedByID:         appCfg.CreatedByID,
		VCSConnectionCommit: appCfg.VCSConnectionCommit,
		Config:              &cfg,
		Resources:           resources,
	})
}

func (s *service) getAppConfigResources(ctx context.Context, appCfg *app.AppConfig) ([]app.AppConfigResource, error) {
	resources := []app.AppConfigResource{}

	lookups := []struct {
		kind  app.AppConfigResourceKind
		model any
		ids   []string
	}{
		{app.AppConfigResourceKindComponent, &app.Component{}, appCfg.ComponentIDs},
		{app.AppConfigResourceKindAction, &app.ActionWorkflow{}, appCfg.ActionIDs},
		{app.AppConfigResourceKindRunbook, &app.Runbook{}, appCfg.RunbookIDs},
	}

	for _, lookup := range lookups {
		if len(lookup.ids) == 0 {
			continue
		}

		var rows []struct {
			ID   string
			Name string
		}
		if err := s.db.WithContext(ctx).
			Model(lookup.model).
			Select("id", "name").
			Where("id IN ?", lookup.ids).
			Find(&rows).Error; err != nil {
			return nil, fmt.Errorf("unable to get %ss: %w", lookup.kind, err)
		}

		for _, row := range rows {
			resources = append(resources, app.AppConfigResource{
				ID:   row.ID,
				Name: row.Name,
				Kind: lookup.kind,
			})
		}
	}

	return resources, nil
}
