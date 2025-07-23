package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

// DEPRECATED: This endpoint is deprecated and will be removed in a future release.

// @ID						DeleteInstall
// @Summary				delete an install
// @Description.markdown	delete_install.md
// @Param					install_id	path	string	true	"install ID"
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
// @Success				200	{boolean}	true
// @Router					/v1/installs/{install_id} [DELETE]
func (s *service) DeleteInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	workflow, err := s.helpers.CreateWorkflow(ctx,
		install.ID,
		app.WorkflowTypeDeprovision,
		map[string]string{},
		app.StepErrorBehaviorAbort,
		false,
	)
	if err != nil {
		ctx.Error(err)
		return
	}
	s.evClient.Send(ctx, install.ID, &signals.Signal{
		Type:              signals.OperationExecuteFlow,
		InstallWorkflowID: workflow.ID,
	})

	s.evClient.Send(ctx, install.ID, &signals.Signal{
		Type: signals.OperationForget,
	})

	ctx.Header(app.HeaderInstallWorkflowID, workflow.ID)

	ctx.JSON(http.StatusOK, true)
}
