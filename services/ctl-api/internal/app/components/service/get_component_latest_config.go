package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetAppComponentLatestConfig
// @Summary				get latest config for a component
// @Description.markdown	get_component_latest_config.md
// @Param					app_id			path	string	true	"app ID"
// @Param					component_id	path	string	true	"component ID"
// @Tags					components
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Accept					json
// @Produce				json
// @Success				200	{object}	app.ComponentConfigConnection
// @Router					/v1/apps/{app_id}/components/{component_id}/configs/latest [GET]
func (s *service) GetAppComponentLatestConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	cmpID := ctx.Param("component_id")

	comp, err := s.helpers.GetComponent(ctx, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component configs: %w", err))
		return
	}

	// Validate component belongs to the requesting org AND the specified app
	if comp.OrgID != org.ID || comp.AppID != appID {
		ctx.Error(fmt.Errorf("component not found: %w", gorm.ErrRecordNotFound))
		return
	}

	if comp.LatestConfig == nil {
		ctx.Error(fmt.Errorf("no config found for component: %w", gorm.ErrRecordNotFound))
		return
	}

	ctx.JSON(http.StatusOK, comp.LatestConfig)
}

// @ID						GetComponentLatestConfig
// @Summary				get latest config for a component
// @Description.markdown	get_component_latest_config.md
// @Param					component_id	path	string	true	"component ID"
// @Tags					components
// @Security				APIKey
// @Security				OrgID
// @Deprecated    true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Accept					json
// @Produce				json
// @Success				200	{object}	app.ComponentConfigConnection
// @Router					/v1/components/{component_id}/configs/latest [GET]
func (s *service) GetComponentLatestConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	cmpID := ctx.Param("component_id")

	comp, err := s.helpers.GetComponent(ctx, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component configs: %w", err))
		return
	}

	// Validate component belongs to the requesting org
	if comp.OrgID != org.ID {
		ctx.Error(fmt.Errorf("component not found: %w", gorm.ErrRecordNotFound))
		return
	}

	if comp.LatestConfig == nil {
		ctx.Error(fmt.Errorf("no config found for component: %w", gorm.ErrRecordNotFound))
		return
	}

	ctx.JSON(http.StatusOK, comp.LatestConfig)
}
