package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// @ID GetInstallActionWorkflowRun
// @Summary	get action workflow runs by install id and run id
// @Description.markdown	get_install_action_workflow_runs.md
// @Param			install_id	path	string	true	"install ID"
// @Param			run_id	path	string	true	"run ID"
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
// @Success		200				{object}	app.InstallActionWorkflowRun
// @Router			/v1/installs/{install_id}/action-workflows/runs/{run_id} [get]
func (s *service) GetInstallActionWorkflowRun(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	runID := ctx.Param("run_id")
	configs, err := s.findInstallActionWorkflowRun(ctx, org.ID, runID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install action workflow run by id %s: %w", runID, err))
		return
	}

	ctx.JSON(http.StatusOK, configs)
}

func (s *service) findInstallActionWorkflowRun(ctx context.Context, orgID, runID string) (*app.InstallActionWorkflowRun, error) {
	runs := &app.InstallActionWorkflowRun{}
	res := s.db.WithContext(ctx).
		Where("org_id = ? AND id = ?", orgID, runID).
		Find(&runs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install action workflow runs: %w", res.Error)
	}

	return runs, nil
}
