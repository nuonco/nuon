package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetAppActionConfig
// @Summary				get an app action config
// @Description.markdown	get_action_workflow_config.md
// @Param					app_id				path	string	true	"app ID"
// @Param					action_id			path	string	true	"action ID"
// @Param					action_config_id	path	string	true	"action config ID"
// @Tags					actions,actions/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.ActionWorkflowConfig
// @Router					/v1/apps/{app_id}/actions/{action_id}/configs/{action_workflow_config_id} [get]
func (s *service) GetAppActionConfig(ctx *gin.Context) {
	awcID := ctx.Param("action_workflow_config_id")
	awc, err := s.findActionWorkflowConfig(ctx, awcID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get action workflow config %s: %w", awcID, err))
		return
	}

	ctx.JSON(http.StatusOK, awc)
}

// @ID						GetActionWorkflowConfig
// @Summary				get an app action workflow config
// @Description.markdown	get_action_workflow_config.md
// @Param					action_workflow_config_id	path	string	true	"action workflow config ID"
// @Tags					actions,actions/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated  			true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.ActionWorkflowConfig
// @Router					/v1/action-workflows/configs/{action_workflow_config_id} [get]
func (s *service) GetActionWorkflowConfig(ctx *gin.Context) {
	awcID := ctx.Param("action_workflow_config_id")
	awc, err := s.findActionWorkflowConfig(ctx, awcID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get action workflow config %s: %w", awcID, err))
		return
	}

	ctx.JSON(http.StatusOK, awc)
}

func (s *service) findActionWorkflowConfig(ctx context.Context, awcID string) (*app.ActionWorkflowConfig, error) {
	aw := app.ActionWorkflowConfig{}
	res := s.db.WithContext(ctx).
		Preload("Triggers").
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("action_workflow_step_configs.idx ASC")
		}).
		Where("id = ?", awcID).
		First(&aw)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get action workflow config: %w", res.Error)
	}
	return &aw, nil
}
