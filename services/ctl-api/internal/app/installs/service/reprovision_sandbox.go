package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

type ReprovisionInstallSandboxRequest struct {
	PlanOnly bool `json:"plan_only"`
}

// @ID						ReprovisionInstallSandbox
// @Summary				reprovision an install sandbox
// @Description.markdown	reprovision_install_sandbox.md
// @Param					install_id	path	string						true	"install ID"
// @Param					req			body	ReprovisionInstallSandboxRequest	true	"Input"
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
// @Router					/v1/installs/{install_id}/reprovision-sandbox [post]
func (s *service) ReprovisionInstallSandbox(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req ReprovisionInstallSandboxRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	// TODO(jm): remove this once the legacy install flow is deprecated
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureIndependentRunner)
	if err != nil {
		ctx.Error(err)
		return
	}
	if !enabled {
		ctx.JSON(http.StatusCreated, install)
		return
	}

	workflow, err := s.helpers.CreateInstallFlow(ctx,
		install.ID,
		app.WorkflowTypeReprovisionSandbox,
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
