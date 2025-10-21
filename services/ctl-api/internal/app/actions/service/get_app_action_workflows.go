package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetAppActions
// @Summary				get action workflows for an app
// @Description.markdown	get_app_action_workflows.md
// @Param					app_id						path	string	true	"app ID"
// @Param         q 						query	string	false	"search query to filter action workflows by name"
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
// @Success				200	{array}		app.ActionWorkflow
// @Router					/v1/apps/{app_id}/action-workflows [get]
func (s *service) GetAppActions(ctx *gin.Context) {
	s.GetAppActionWorkflows(ctx)
}

// @ID						GetActionWorkflows
// @Summary				get action workflows for an app
// @Description.markdown	get_app_action_workflows.md
// @Param					app_id						path	string	true	"app ID"
// @Param         q 						query	string	false	"search query to filter action workflows by name"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					actions
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
// @Success				200	{array}		app.ActionWorkflow
// @Router					/v1/apps/{app_id}/action-workflows [get]
func (s *service) GetAppActionWorkflows(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	q := ctx.Query("q")
	appID := ctx.Param("app_id")
	_, err = s.findApp(ctx, org.ID, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app %s: %w", appID, err))
		return
	}

	actionWorkflows, err := s.findActionWorkflows(ctx, org.ID, appID, q)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get action workflows %s: %w", appID, err))
		return
	}

	ctx.JSON(http.StatusOK, actionWorkflows)
}

func (s *service) findActionWorkflows(ctx *gin.Context, orgID, appID, q string) ([]*app.ActionWorkflow, error) {
	actionWorkflows := []*app.ActionWorkflow{}
	tx := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Preload("Configs", func(db *gorm.DB) *gorm.DB {
			return db.Scopes(scopes.WithOverrideTable("action_workflow_configs_latest_view_v1"))
		}).
		Preload("Configs.Triggers").
		Preload("Configs.Triggers.Component").
		Preload("Configs.Steps").
		Where("org_id = ? AND app_id = ?", orgID, appID)

	if q != "" {
		tx = tx.Where("name ILIKE ?", "%"+q+"%")
	}
	res := tx.Find(&actionWorkflows)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get action workflows: %w", res.Error)
	}

	actionWorkflows, err := db.HandlePaginatedResponse(ctx, actionWorkflows)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return actionWorkflows, nil
}

func (s *service) findApp(ctx context.Context, orgID, appID string) (*app.App, error) {
	app := app.App{}
	res := s.db.WithContext(ctx).
		Preload("Org").
		Preload("Org.VCSConnections").
		Where("name = ? AND org_id = ?", appID, orgID).
		Or("id = ?", appID).
		First(&app)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	return &app, nil
}
