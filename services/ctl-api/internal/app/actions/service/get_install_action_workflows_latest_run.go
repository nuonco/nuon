package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type ActionWorkflowLatestRunResponse struct {
	ActionWorkFlow app.ActionWorkflow            `json:"action_workflow"`
	LatestRun      *app.InstallActionWorkflowRun `json:"install_action_workflow_run"`
}

// @ID GetInstallActionWorkflowsLatestRun
// @Summary	get latest runs for all action workflows by install id
// @Description.markdown	get_install_action_workflows_latest_run.md
// @Param			install_id	path	string	true	"install ID"
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
// @Success		200				{array}	ActionWorkflowLatestRunResponse
// @Router			/v1/installs/{install_id}/action-workflows/latest-runs [get]
func (s *service) GetInstallActionWorkflowsLatestRun(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")
	install, err := s.findInstall(ctx, org.ID, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}

	actionWorkflows, err := s.findActionWorkflows(ctx, org.ID, install.AppID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get action workflows %s: %w", install.AppID, err))
		return
	}

	if len(actionWorkflows) == 0 {
		ctx.JSON(http.StatusOK, []ActionWorkflowLatestRunResponse{})
	}

	// init response
	var response []ActionWorkflowLatestRunResponse

	for _, actionWorkflow := range actionWorkflows {
		run, err := s.findLatestInstallActionWorkflowRun(ctx, org.ID, installID, actionWorkflow.ID)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to get install action workflow runs %s: %w", installID, err))
			return
		}

		response = append(response, ActionWorkflowLatestRunResponse{
			ActionWorkFlow: *actionWorkflow,
			LatestRun:      run,
		})
	}

	ctx.JSON(http.StatusOK, response)
}

func (s *service) findLatestInstallActionWorkflowRun(ctx context.Context, orgID, installID, actionWorkflowID string) (*app.InstallActionWorkflowRun, error) {
	var run app.InstallActionWorkflowRun
	res := s.db.WithContext(ctx).
		Preload("RunnerJob").
		Joins("JOIN action_workflow_configs ON action_workflow_configs.id = install_action_workflow_runs.action_workflow_config_id").
		Where("install_action_workflow_runs.org_id = ? AND install_action_workflow_runs.install_id = ? and action_workflow_configs.action_workflow_id = ?", orgID, installID, actionWorkflowID).
		Order("created_at desc").
		First(&run)

	if res.Error != nil && errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install action workflow run: %w", res.Error)
	}

	return &run, nil
}
