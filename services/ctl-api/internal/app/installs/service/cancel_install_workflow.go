package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

// @ID						CancelInstallWorkflow
// @Summary				cancel an ongoing install workflow
// @Description.markdown	cancel_install_workflow.md
// @Param install_workflow_id	path	string true "install workflow ID"
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
// @Success				202	{boolean}		true
// @Router					/v1/install-workflows/{install_workflow_id}/cancel [post]
func (s *service) CancelInstallWorkflow(ctx *gin.Context) {
	workflowID := ctx.Param("install_workflow_id")

	wf, err := s.getInstallWorkflow(ctx, workflowID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install workflow: %w", err))
		return
	}

	// if wf.Status != app.InstallWorkflowStatusRunning {
	// 	ctx.Error(fmt.Errorf("install workflow is not running"))
	// 	return
	// }

	id := fmt.Sprintf("event-loop-%s-execute-workflow-steps", wf.InstallID)

	err = s.evClient.Cancel(ctx, signals.TemporalNamespace, id)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to cancel install workflow: %w", err))
		return
	}

	ctx.JSON(http.StatusAccepted, true)
}
