package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetAppOperationRoleConfigs
// @Summary				get operation role configs
// @Description			Get all operation role configs for an app
// @Tags					apps
// @Accept					json
// @Param					app_id	path	string	true	"app ID"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.AppOperationRoleConfig
// @Router					/v1/apps/{app_id}/operation-role-configs [get]
func (s *service) GetAppOperationRoleConfigs(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")

	// Verify app exists and belongs to org
	var appEntity app.App
	res := s.db.WithContext(ctx).Where("id = ? AND org_id = ?", appID, org.ID).First(&appEntity)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app: %w", res.Error))
		return
	}

	// Get all app configs for this app
	var appConfigs []app.AppConfig
	res = s.db.WithContext(ctx).Where("app_id = ?", appID).Find(&appConfigs)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app configs: %w", res.Error))
		return
	}

	// Extract app config IDs
	appConfigIDs := make([]string, len(appConfigs))
	for i, cfg := range appConfigs {
		appConfigIDs[i] = cfg.ID
	}

	// Get operation role configs for these app configs
	var configs []app.AppOperationRoleConfig
	res = s.db.WithContext(ctx).
		Preload("Rules").
		Where("app_config_id IN ?", appConfigIDs).
		Find(&configs)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find operation role configs: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, configs)
}
