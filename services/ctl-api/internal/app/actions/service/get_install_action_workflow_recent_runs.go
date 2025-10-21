package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetInstallActionRecentRuns
// @Summary				get recent runs for an action workflow by install id
// @Description.markdown	get_install_action_workflow_recent_runs.md
// @Param					install_id					path	string	true	"install ID"
// @Param					action_id			path	string	true	"action workflow ID"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					actions
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated     true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.InstallActionWorkflow
// @Router					/v1/installs/{install_id}/actions/{action_id}/recent-runs [get]
func (s *service) GetInstallActionRecentRuns(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")
	actionWorkflowID := ctx.Param("action_id")
	iaw, err := s.getRecentRuns(ctx, org.ID, installID, actionWorkflowID)

	ctx.JSON(http.StatusOK, iaw)
}

// @ID						GetInstallActionWorkflowRecentRuns
// @Summary				get recent runs for an action workflow by install id
// @Description.markdown	get_install_action_workflow_recent_runs.md
// @Param					install_id					path	string	true	"install ID"
// @Param					action_workflow_id			path	string	true	"action workflow ID"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					actions
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated     true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.InstallActionWorkflow
// @Router					/v1/installs/{install_id}/action-workflows/{action_workflow_id}/recent-runs [get]
func (s *service) GetInstallActionWorkflowRecentRuns(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")
	actionWorkflowID := ctx.Param("action_workflow_id")
	iaw, err := s.getRecentRuns(ctx, org.ID, installID, actionWorkflowID)

	ctx.JSON(http.StatusOK, iaw)
}

func (s *service) findInstall(ctx context.Context, orgID, installID string) (*app.Install, error) {
	fmt.Println("orgID", orgID)
	install := app.Install{}
	res := s.db.WithContext(ctx).
		Where("id = ? and org_id = ?", installID, orgID).
		First(&install)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	return &install, nil
}

func (s *service) getRecentRuns(ctx *gin.Context, orgID, installID, actionWorkflowID string) (*app.InstallActionWorkflow, error) {
	var installActionWorkflow app.InstallActionWorkflow
	// TODO remove reading from query params
	limitStr := ctx.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return nil, fmt.Errorf("unable to convert limit to int: %w", err)
	}

	res := s.db.WithContext(ctx).
		Where(app.InstallActionWorkflow{
			InstallID:        installID,
			ActionWorkflowID: actionWorkflowID,
			OrgID:            orgID,
		}).
		Preload("ActionWorkflow").
		Preload("ActionWorkflow.Configs", func(db *gorm.DB) *gorm.DB {
			return db.
				Order("action_workflow_configs.created_at DESC").
				Limit(1)
		}).
		Preload("ActionWorkflow.Configs.Triggers").
		Preload("ActionWorkflow.Configs.Steps").
		Preload("ActionWorkflow.Configs.Steps.PublicGitVCSConfig").
		Preload("ActionWorkflow.Configs.Steps.ConnectedGithubVCSConfig").
		Preload("Runs", func(db *gorm.DB) *gorm.DB {
			return db.
				Scopes(scopes.WithOffsetPagination).
				Preload("CreatedBy").
				Order("install_action_workflow_runs.created_at DESC").
				Limit(limit)
		}).
		First(&installActionWorkflow)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install action workflow")
	}

	installActionWorkflow.Runs, err = db.HandlePaginatedResponse(ctx, installActionWorkflow.Runs)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return &installActionWorkflow, nil
}
