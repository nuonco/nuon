package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type UpdateInstallActionWorkflowRunStepRequest struct {
	Status            app.InstallActionWorkflowRunStepStatus `json:"status"`
	ExecutionDuration time.Duration                          `json:"execution_duration" swaggertype:"primitive,integer"`
}

// @ID						UpdateInstallActionWorkflowRunStep
// @Summary				update an action workflow run step by
// @Description.markdown	update_install_action_workflow_run_step.md
// @Param					install_id		path	string										true	"install ID"
// @Param					workflow_run_id	path	string										true	"workflow run ID"
// @Param					step_id			path	string										true	"step ID"
// @Param					req				body	UpdateInstallActionWorkflowRunStepRequest	true	"Input"
// @Tags					actions/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.InstallActionWorkflowRunStep
// @Router					/v1/installs/{install_id}/action-workflow-runs/{workflow_run_id}/steps/{step_id} [PUT]
func (s *service) UpdateInstallActionWorkflowRunStep(ctx *gin.Context) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	installID := ctx.Param("install_id")
	stepID := ctx.Param("step_id")

	var req UpdateInstallActionWorkflowRunStepRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	updatedStep, err := s.updateInstallActionWorkflowRunStep(ctx, orgID, installID, stepID, &req)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to update install action workflow run step"))
		return
	}

	ctx.JSON(http.StatusOK, updatedStep)
}

func (s *service) updateInstallActionWorkflowRunStep(ctx context.Context, orgID, installID, stepID string, req *UpdateInstallActionWorkflowRunStepRequest) (*app.InstallActionWorkflowRunStep, error) {
	step := app.InstallActionWorkflowRunStep{
		Status:            req.Status,
		ExecutionDuration: req.ExecutionDuration,
	}

	installRuns := s.db.Model(&app.InstallActionWorkflowRun{}).
		Select("id").
		Where(app.InstallActionWorkflowRun{OrgID: orgID, InstallID: installID})

	res := s.db.WithContext(ctx).
		Model(&app.InstallActionWorkflowRunStep{}).
		Where(app.InstallActionWorkflowRunStep{ID: stepID, OrgID: orgID}).
		Where("install_action_workflow_run_id IN (?)", installRuns).
		Updates(&step)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to update step")
	}
	if res.RowsAffected < 1 {
		return nil, fmt.Errorf("no step found: %s %w", stepID, gorm.ErrRecordNotFound)
	}

	return &step, nil
}
