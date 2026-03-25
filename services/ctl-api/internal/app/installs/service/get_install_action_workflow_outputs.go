package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetInstallActionWorkflowOutputs
// @Summary				get an install action workflow outputs
// @Description.markdown	get_install_action_workflow_outputs.md
// @Param					install_id	path	string	true	"install ID"
// @Param					action_id	path	string	true	"action workflow ID or name"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	map[string]interface{}
// @Router					/v1/installs/{install_id}/actions/{action_id}/outputs [get]
func (s *service) GetInstallActionWorkflowOutputs(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	actionID := ctx.Param("action_id")

	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	outputs, err := s.getInstallActionWorkflowOutputs(ctx, org.ID, installID, actionID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get action workflow outputs %s: %w", actionID, err))
		return
	}

	ctx.JSON(http.StatusOK, outputs)
}

func (s *service) getInstallActionWorkflowOutputs(ctx context.Context, orgID, installID, actionID string) (map[string]interface{}, error) {
	// find the InstallActionWorkflow by ID or name, scoped to org
	var iaw app.InstallActionWorkflow
	res := s.db.WithContext(ctx).
		Joins("JOIN action_workflows ON action_workflows.id = install_action_workflows.action_workflow_id").
		Where("install_action_workflows.org_id = ? AND install_action_workflows.install_id = ? AND (action_workflows.id = ? OR action_workflows.name = ?)",
			orgID, installID, actionID, actionID).
		First(&iaw)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to find install action workflow")
	}

	// find the latest run and its runner job
	var run app.InstallActionWorkflowRun
	res = s.db.WithContext(ctx).
		Preload("RunnerJob").
		Where("install_action_workflow_id = ?", iaw.ID).
		Order("created_at DESC").
		First(&run)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to find action workflow run")
	}

	return run.Outputs, nil
}
