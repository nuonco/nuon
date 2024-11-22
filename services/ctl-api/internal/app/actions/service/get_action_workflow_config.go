package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// @ID GetActionWorkflowConfig
// @Summary	get an app action workflow config
// @Description.markdown	get_action_workflow_config.md
// @Param			action_workflow_config_id	path	string	true	"action workflow config ID"
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
// @Success		200				{object}	app.ActionWorkflowConfig
// @Router			/v1/action-workflows/configs/{action_workflow_config_id} [get]
func (s *service) GetActionWorkflowConfig(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	awcID := ctx.Param("action_workflow_config_id")
	awc, err := s.findActionWorkflowConfig(ctx, org.ID, awcID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app %s: %w", awcID, err))
		return
	}

	ctx.JSON(http.StatusOK, awc)
}

func (s *service) findActionWorkflowConfig(ctx context.Context, orgID, awID string) (*app.ActionWorkflowConfig, error) {
	aw := app.ActionWorkflowConfig{}
	res := s.db.WithContext(ctx).
		Where("org_id = ? AND id = ?", orgID, awID).
		First(&aw)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get action workflow config: %w", res.Error)
	}
	return &aw, nil
}
