package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

// Deprecated: This endpoint is deprecated and will be removed in a future version.

// @ID						DeleteInstallComponent
// @Summary				delete an install component
// @Description.markdown	delete_install_component.md
// @Param					install_id		path	string				true	"install ID"
// @Param					component_id	path	string				true	"component ID"
// @Param					force					query	bool					false	"force delete"
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
// @Success				200	{boolean} 		true
// @Router					/v1/installs/{install_id}/components/{component_id} [delete]
func (s *service) DeleteInstallComponent(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	componentID := ctx.Param("component_id")

	install, err := s.helpers.GetInstall(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install: %w", err))
		return
	}

	component, err := s.helpers.GetComponent(ctx, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get component: %w", err))
		return
	}

	workflow, err := s.helpers.CreateInstallWorkflow(ctx,
		install.ID,
		app.InstallWorkflowTypeTeardownComponent,
		map[string]string{
			"component_id": component.ID,
		},
		app.StepErrorBehaviorAbort,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.evClient.Send(ctx, installID, &signals.Signal{
		Type:              signals.OperationExecuteWorkflow,
		InstallWorkflowID: workflow.ID,
	})

	ctx.Header(app.HeaderInstallWorkflowID, workflow.ID)

	ctx.JSON(http.StatusOK, true)
}
