package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetActionLatestConfig
// @Summary				get an app action workflow's latest config
// @Description.markdown	get_action_workflow_latest_config.md
// @Param					app_id				path	string	true	"app ID"
// @Param					action_workflow_id	path	string	true	"action workflow ID"
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
// @Router					/v1/apps/{app_id}/actions/{action_id}/latest-config [get]
func (s *service) GetAppActionLatestConfig(ctx *gin.Context) {
	awID := ctx.Param("action_id")
	awc, err := s.getActionWorkflowLatestConfig(ctx, awID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get action workflow config %s: %w", awID, err))
		return
	}

	ctx.JSON(http.StatusOK, awc)
}

//		@ID						GetActionWorkflowLatestConfig
//		@Summary				get an app action workflow's latest config
//		@Description.markdown	get_action_workflow_latest_config.md
//		@Param					action_workflow_id	path	string	true	"action workflow ID"
//		@Tags					actions,actions/runner
//		@Accept					json
//		@Produce				json
//		@Security				APIKey
//		@Security				OrgID
//	 @Deprecated  			true
//		@Failure				400	{object}	stderr.ErrResponse
//		@Failure				401	{object}	stderr.ErrResponse
//		@Failure				403	{object}	stderr.ErrResponse
//		@Failure				404	{object}	stderr.ErrResponse
//		@Failure				500	{object}	stderr.ErrResponse
//		@Success				200	{object}	app.ActionWorkflowConfig
//		@Router					/v1/action-workflows/{action_workflow_id}/latest-config [get]
func (s *service) GetActionWorkflowLatestConfig(ctx *gin.Context) {
	awID := ctx.Param("action_workflow_id")
	awc, err := s.getActionWorkflowLatestConfig(ctx, awID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get action workflow config %s: %w", awID, err))
		return
	}

	ctx.JSON(http.StatusOK, awc)
}

func (s *service) getActionWorkflowLatestConfig(ctx context.Context, awcID string) (*app.ActionWorkflowConfig, error) {
	aw := app.ActionWorkflowConfig{}
	res := s.db.WithContext(ctx).
		Preload("Triggers").
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("action_workflow_step_configs.idx ASC")
		}).
		Order("created_at DESC").
		Limit(1).
		First(&aw, "action_workflow_id = ?", awcID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get action workflow latest config: %w", res.Error)
	}

	return &aw, nil
}
