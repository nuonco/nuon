package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type TeardownInstallComponentsRequest struct {
	ErrorBehavior app.StepErrorBehavior `json:"error_behavior" swaggertype:"string"`
	PlanOnly      bool                  `json:"plan_only"`
}

func (c *TeardownInstallComponentsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						TeardownInstallComponents
// @Summary				teardown an install's components
// @Description.markdown	teardown_install_components.md
// @Param					install_id	path	string								true	"install ID"
// @Param					req			body	TeardownInstallComponentsRequest	true	"Input"
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
// @Router					/v1/installs/{install_id}/components/teardown-all [post]
func (s *service) TeardownInstallComponents(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req TeardownInstallComponentsRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	_, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	installCmps, err := s.helpers.GetInstallComponents(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install components: %w", err))
		return
	}

	if len(installCmps) == 0 {
		ctx.JSON(http.StatusNoContent, "no components to teardown")
		return
	}

	allInactive := true
	for _, cmp := range installCmps {
		if cmp.Status != app.InstallComponentStatusInactive {
			allInactive = false
			break
		}
	}
	if allInactive {
		ctx.Error(fmt.Errorf("install components are already inactive"))
		return
	}

	workflow, err := s.helpers.CreateWorkflow(ctx,
		installID,
		app.WorkflowTypeTeardownComponents,
		map[string]string{},
		req.ErrorBehavior,
		req.PlanOnly,
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
