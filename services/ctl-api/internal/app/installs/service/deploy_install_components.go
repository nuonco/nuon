package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

type DeployInstallComponentsRequest struct {
	ErrorBehavior app.StepErrorBehavior `json:"error_behavior" swaggertype:"string"`
}

func (c *DeployInstallComponentsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID						DeployInstallComponents
// @Summary				deploy all components on an install
// @Description.markdown	install_deploy_components.md
// @Param					install_id	path	string							true	"install ID"
// @Param					req			body	DeployInstallComponentsRequest	true	"Input"
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

	var req DeployInstallComponentsRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	_, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureIndependentRunner)
	if err != nil {
		ctx.Error(err)
		return
	}
	if !enabled {
		s.evClient.Send(ctx, installID, &signals.Signal{
			Type: signals.OperationDeployComponents,
		})

		ctx.JSON(http.StatusCreated, "ok")
		return
	}

	workflow, err := s.helpers.CreateInstallWorkflow(ctx,
		installID,
		app.InstallWorkflowTypeDeployComponents,
		map[string]string{},
		req.ErrorBehavior)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.evClient.Send(ctx, installID, &signals.Signal{
		Type:              signals.OperationExecuteWorkflow,
		InstallWorkflowID: workflow.ID,
	})

	ctx.JSON(http.StatusCreated, "ok")
}
