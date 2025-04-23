package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

type DeprovisionInstallSandboxRequest struct {
	ErrorBehavior app.StepErrorBehavior `json:"error_behavior" swaggertype:"string"`
}

func (c *DeprovisionInstallSandboxRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID              DeprovisionInstallSandbox
// @Summary         deprovision an install
// @Description.markdown deprovision_install_sandbox.md
// @Param           install_id path string true "install ID"
// @Param           req body DeprovisionInstallSandboxRequest true "Input"
// @Tags            installs
// @Accept          json
// @Produce         json
// @Security        APIKey
// @Security        OrgID
// @Failure         400 {object} stderr.ErrResponse
// @Failure         401 {object} stderr.ErrResponse
// @Failure         403 {object} stderr.ErrResponse
// @Failure         404 {object} stderr.ErrResponse
// @Failure         500 {object} stderr.ErrResponse
// @Success         201 {string} ok
// @Router          /v1/installs/{install_id}/deprovision-sandbox [post]
func (s *service) DeprovisionInstallSandbox(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req DeprovisionInstallSandboxRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	workflow, err := s.helpers.CreateInstallWorkflow(ctx,
		install.ID,
		app.InstallWorkflowTypeDeprovisionSandbox,
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
