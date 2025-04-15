package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

type ReprovisionInstallRequest struct {
	ErrorBehavior app.StepErrorBehavior `json:"error_behavior" swaggertype:"string"`
}

func (c *ReprovisionInstallRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID						ReprovisionInstall
// @Summary				reprovision an install
// @Description.markdown	reprovision_install.md
// @Param					install_id	path	string						true	"install ID"
// @Param					req			body	ReprovisionInstallRequest	true	"Input"
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

	var req ReprovisionInstallRequest
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
		s.evClient.Send(ctx, install.ID, &signals.Signal{
			Type: signals.OperationReprovision,
		})
		ctx.JSON(http.StatusCreated, install)
		return
	}

	workflow, err := s.helpers.CreateInstallWorkflow(ctx,
		install.ID,
		app.InstallWorkflowTypeReprovision,
		map[string]string{},
		req.ErrorBehavior)
	if err != nil {
		ctx.Error(err)
		return
	}
	s.evClient.Send(ctx, install.ID, &signals.Signal{
		Type:              signals.OperationExecuteWorkflow,
		InstallWorkflowID: workflow.ID,
	})

	ctx.JSON(http.StatusCreated, "ok")
}
