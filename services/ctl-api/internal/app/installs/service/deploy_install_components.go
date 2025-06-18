package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

type DeployInstallComponentsRequest struct{}

// @ID						DeployInstallComponents
// @Summary				deploy all components on an install
// @Description.markdown	install_deploy_components.md
// @Param					install_id	path	string							true	"install ID"
// @Param					req			body	DeployInstallComponentsRequest	false	"Input"
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
// @Success				201	{string}	ok
// @Router					/v1/installs/{install_id}/components/deploy-all [post]
func (s *service) DeployInstallComponents(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	_, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	workflow, err := s.helpers.CreateInstallFlow(ctx,
		installID,
		app.InstallWorkflowTypeDeployComponents,
		map[string]string{},
		app.StepErrorBehaviorAbort,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.evClient.Send(ctx, installID, &signals.Signal{
		Type:              signals.OperationExecuteFlow,
		InstallWorkflowID: workflow.ID,
	})

	ctx.Header(app.HeaderInstallWorkflowID, workflow.ID)

	ctx.JSON(http.StatusCreated, "ok")
}
