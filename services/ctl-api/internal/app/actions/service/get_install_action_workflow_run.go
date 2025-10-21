package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

//		@ID						GetInstallActionRun
//		@Summary				get action workflow runs by install id and run id
//		@Description.markdown	get_install_action_workflow_run.md
//		@Param					install_id	path	string	true	"install ID"
//		@Param					run_id		path	string	true	"run ID"
//		@Tags					actions, actions/runner
//		@Accept					json
//		@Produce				json
//		@Security				APIKey
//		@Security				OrgID
//	 @Deprecated     true
//		@Failure				400	{object}	stderr.ErrResponse
//		@Failure				401	{object}	stderr.ErrResponse
//		@Failure				403	{object}	stderr.ErrResponse
//		@Failure				404	{object}	stderr.ErrResponse
//		@Failure				500	{object}	stderr.ErrResponse
//		@Success				200	{object}	app.InstallActionWorkflowRun
//		@Router					/v1/installs/{install_id}/actions/runs/{run_id} [get]
func (s *service) GetInstallActionRun(ctx *gin.Context) {
	s.GetInstallActionWorkflowRun(ctx)
}

//		@ID						GetInstallActionWorkflowRun
//		@Summary				get action workflow runs by install id and run id
//		@Description.markdown	get_install_action_workflow_run.md
//		@Param					install_id	path	string	true	"install ID"
//		@Param					run_id		path	string	true	"run ID"
//		@Tags					actions, actions/runner
//		@Accept					json
//		@Produce				json
//		@Security				APIKey
//		@Security				OrgID
//	 @Deprecated     true
//		@Failure				400	{object}	stderr.ErrResponse
//		@Failure				401	{object}	stderr.ErrResponse
//		@Failure				403	{object}	stderr.ErrResponse
//		@Failure				404	{object}	stderr.ErrResponse
//		@Failure				500	{object}	stderr.ErrResponse
//		@Success				200	{object}	app.InstallActionWorkflowRun
//		@Router					/v1/installs/{install_id}/action-workflows/runs/{run_id} [get]
func (s *service) GetInstallActionWorkflowRun(ctx *gin.Context) {
	runID := ctx.Param("run_id")
	run, err := s.findInstallActionWorkflowRun(ctx, runID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install action workflow run by id %s: %w", runID, err))
		return
	}

	ctx.JSON(http.StatusOK, run)
}

func (s *service) findInstallActionWorkflowRun(ctx context.Context, runID string) (*app.InstallActionWorkflowRun, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	run := &app.InstallActionWorkflowRun{}
	res := s.db.WithContext(ctx).
		Preload("ActionWorkflowConfig").
		Preload("ActionWorkflowConfig.Steps").
		Preload("ActionWorkflowConfig.Triggers").
		Preload("LogStream").
		Preload("RunnerJob").
		Preload("Steps").
		Where("org_id = ? AND id = ?", orgID, runID).
		First(&run)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install action workflow runs: %w", res.Error)
	}

	return run, nil
}
