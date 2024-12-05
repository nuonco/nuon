package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// @ID GetActionWorkflowConfigs
// @Summary	get action workflow for an app
// @Description.markdown	get_action_workflow_configs.md
// @Param			action_workflow_id	path	string	true	"action workflow ID"
// @Tags			actions
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{array}	app.ActionWorkflowConfig
// @Router			/v1/action-workflows/{action_workflow_id}/configs [get]
func (s *service) GetActionWorkflowConfigs(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	awID := ctx.Param("action_workflow_id")
	configs, err := s.findActionWorkflowConfigs(ctx, org.ID, awID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get action workflow %s: %w", awID, err))
		return
	}

	ctx.JSON(http.StatusOK, configs)
}

func (s *service) findActionWorkflowConfigs(ctx context.Context, orgID, awID string) ([]*app.ActionWorkflowConfig, error) {
	actionWorkflowConfigs := []*app.ActionWorkflowConfig{}
	res := s.db.WithContext(ctx).
		Preload("Triggers").
		Preload("Steps").
		Where("org_id = ? AND action_workflow_id = ?", orgID, awID).
		Find(&actionWorkflowConfigs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get action workflow configs: %w", res.Error)
	}

	return actionWorkflowConfigs, nil
}
