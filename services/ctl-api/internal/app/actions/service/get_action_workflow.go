package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetActionWorkflow
// @Summary				get an app action workflow by action workflow id
// @Description.markdown	get_app_action_workflow.md
// @Param					action_workflow_id	path	string	true	"action workflow ID or name"
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
// @Success				200	{object}	app.ActionWorkflow
// @Router					/v1/action-workflows/{action_workflow_id} [get]
func (s *service) GetActionWorkflow(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	awID := ctx.Param("action_workflow_id")
	aw, err := s.findActionWorkflow(ctx, org.ID, awID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app %s: %w", awID, err))
		return
	}

	ctx.JSON(http.StatusOK, aw)
}

func (s *service) findActionWorkflow(ctx context.Context, orgID, awID string) (*app.ActionWorkflow, error) {
	aw := app.ActionWorkflow{}
	res := s.db.WithContext(ctx).
		Preload("Configs", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(scopes.WithOverrideTable("action_workflow_configs_latest_view_v1"))
		}).
		Preload("Configs.Triggers").
		Preload("Configs.Triggers.Component").
		Preload("Configs.Steps").
		Preload("Configs.Steps.PublicGitVCSConfig").
		Preload("Configs.Steps.ConnectedGithubVCSConfig").
		Where("org_id = ? AND id = ?", orgID, awID).
		Or("org_id = ? AND name = ?", orgID, awID).
		First(&aw)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get action workflow: %w", res.Error)
	}

	return &aw, nil
}
