package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

type ReprovisionInstallRequest struct {
	PlanOnly bool `json:"plan_only"`
}

// @ID						ReprovisionInstall
// @Summary				reprovision an install
// @Description.markdown	reprovision_install.md
// @Param					install_id	path	string						true	"install ID"
// @Param					req			body	ReprovisionInstallRequest	false	"Input"
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
// @Router					/v1/installs/{install_id}/reprovision [post]
func (s *service) ReprovisionInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req ReprovisionInstallRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	workflow, err := s.helpers.CreateInstallFlow(ctx,
		install.ID,
		app.WorkflowTypeReprovision,
		map[string]string{},
		app.StepErrorBehaviorAbort,
		req.PlanOnly,
	)
	if err != nil {
		ctx.Error(err)
		return
	}
	s.evClient.Send(ctx, install.ID, &signals.Signal{
		Type:              signals.OperationExecuteFlow,
		InstallWorkflowID: workflow.ID,
	})

	ctx.Header(app.HeaderInstallWorkflowID, workflow.ID)

	ctx.JSON(http.StatusCreated, "ok")
}
