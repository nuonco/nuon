package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetActionWorkflowConfigs
// @Summary				get action workflow for an app
// @Description.markdown	get_action_workflow_configs.md
// @Param					action_workflow_id			path	string	true	"action workflow ID"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					actions
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.ActionWorkflowConfig
// @Router					/v1/action-workflows/{action_workflow_id}/configs [get]
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

func (s *service) findActionWorkflowConfigs(ctx *gin.Context, orgID, awID string) ([]*app.ActionWorkflowConfig, error) {
	actionWorkflowConfigs := []*app.ActionWorkflowConfig{}
	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Preload("Triggers").
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("action_workflow_step_configs.idx ASC")
		}).
		Where("org_id = ? AND action_workflow_id = ?", orgID, awID).
		Find(&actionWorkflowConfigs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get action workflow configs: %w", res.Error)
	}

	actionWorkflowConfigs, err := db.HandlePaginatedResponse(ctx, actionWorkflowConfigs)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return actionWorkflowConfigs, nil
}
