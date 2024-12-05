package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// @ID GetActionWorkflows
// @Summary	get action workflow for an app
// @Description.markdown	get_app_action_workflows.md
// @Param			app_id	path	string	true	"app ID"
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
// @Success		200				{array}	app.ActionWorkflow
// @Router	  /v1/apps/{app_id}/action-workflows [get]
func (s *service) GetAppActionWorkflows(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	_, err = s.findApp(ctx, org.ID, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app %s: %w", appID, err))
		return
	}

	actionWorkflows, err := s.findActionWorkflows(ctx, org.ID, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get action workflows %s: %w", appID, err))
		return
	}

	ctx.JSON(http.StatusOK, actionWorkflows)
}

func (s *service) findActionWorkflows(ctx context.Context, orgID, appID string) ([]*app.ActionWorkflow, error) {
	actionWorkflows := []*app.ActionWorkflow{}
	res := s.db.WithContext(ctx).
		Preload("Configs").
		Preload("Configs.Triggers").
		Preload("Configs.Steps").
		Where("org_id = ? AND app_id = ?", orgID, appID).
		Find(&actionWorkflows)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get action workflows: %w", res.Error)
	}
	return actionWorkflows, nil
}

func (s *service) findApp(ctx context.Context, orgID, appID string) (*app.App, error) {
	app := app.App{}
	res := s.db.WithContext(ctx).
		Where("name = ? AND org_id = ?", appID, orgID).
		Or("id = ?", appID).
		First(&app)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return &app, nil
}
