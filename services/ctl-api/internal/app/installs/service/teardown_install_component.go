package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
)

type TeardownInstallComponentRequest struct {
	ErrorBehavior app.StepErrorBehavior `json:"error_behavior" swaggertype:"string"`
}

func (c *TeardownInstallComponentRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID						TeardownInstallComponent
// @Summary				teardown an install component
// @Description.markdown	teardown_install_component.md
// @Param					install_id		path	string							true	"install ID"
// @Param					component_id	path	string							true	"component ID"
// @Param					req				body	TeardownInstallComponentRequest	false	"Input"
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
// @Router					/v1/installs/{install_id}/components/{component_id}/teardown [post]
func (s *service) TeardownInstallComponent(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	componentID := ctx.Param("component_id")

	var req TeardownInstallComponentRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

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

	installComponent, err := s.helpers.GetInstallComponent(ctx, installID, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install component: %w", err))
		return
	}

	if installComponent.Status == app.InstallComponentStatusInactive {
		ctx.Error(fmt.Errorf("install component is already inactive"))
		return
	}

	workflow, err := s.helpers.CreateInstallWorkflow(ctx,
		install.ID,
		app.InstallWorkflowTypeTeardownComponent,
		map[string]string{
			"component_id": component.ID,
		},
		req.ErrorBehavior,
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

	ctx.JSON(http.StatusCreated, "ok")
}
