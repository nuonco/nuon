package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// @ID GetInstallActionWorkflowRecentRuns
// @Summary	get recent runs for an action workflow by install id
// @Description.markdown	get_install_action_workflow_recent_runs.md
// @Param			install_id	path	string	true	"install ID"
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
// @Success		200				{object}	app.InstallActionWorkflow
// @Router			/v1/installs/{install_id}/action-workflows/{action_workflow_id}/recent-runs [get]
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

func (s *service) getRecentRuns(ctx context.Context, orgID, installID, actionWorkflowID string) (*app.InstallActionWorkflow, error) {
	var installActionWorkflow app.InstallActionWorkflow
	res := s.db.WithContext(ctx).
		Where(app.InstallActionWorkflow{
			InstallID:        installID,
			ActionWorkflowID: actionWorkflowID,
			OrgID:            orgID,
		}).
		Preload("ActionWorkflow").
		Preload("ActionWorkflow.Configs").
		Preload("ActionWorkflow.Configs.Triggers").
		Preload("ActionWorkflow.Configs.Steps").
		Preload("ActionWorkflow.Configs.Steps.PublicGitVCSConfig").
		Preload("ActionWorkflow.Configs.Steps.ConnectedGithubVCSConfig").
		Preload("Runs", func(db *gorm.DB) *gorm.DB {
			return db.Limit(50).Order("install_action_workflow_runs.created_at DESC")
		}).
		First(&installActionWorkflow)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get install action workflow")
	}

	return &installActionWorkflow, nil
}
