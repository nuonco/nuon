package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
func (s *service) GetAppComponentLatestConfig(ctx *gin.Context) (any, error) {
	return s.GetComponentLatestConfig(ctx)
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
func (s *service) GetComponentLatestConfig(ctx *gin.Context) (any, error) {
	comp, err := s.readComponent(ctx.Request.Context(), "component_id")
	if err != nil {
		return nil, fmt.Errorf("unable to get component configs: %w", err)
	}

	if comp.LatestConfig == nil {
		return nil, fmt.Errorf("no config found for component: %w", gorm.ErrRecordNotFound)
	}

	return comp.LatestConfig, nil
}
