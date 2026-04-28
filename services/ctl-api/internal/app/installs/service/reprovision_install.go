package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
)

type ReprovisionInstallRequest struct {
	PlanOnly bool   `json:"plan_only"`
	Role     string `json:"role"`
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
// @Success				201	{object}	app.WorkflowResponse
// @Router					/v1/installs/{install_id}/reprovision [post]
func (s *service) ReprovisionInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req ReprovisionInstallRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	workflow, err := s.helpers.CreateWorkflowWithRole(ctx,
		install.ID,
		app.WorkflowTypeReprovision,
		map[string]string{},
		req.PlanOnly,
		req.Role,
	)
	if err != nil {
		ctx.Error(err)
		return
	}
	useQueues, err := s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
	if err != nil {
		ctx.Error(fmt.Errorf("checking features: %w", err))
		return
	}
	if useQueues {
		queueID, err := s.getInstallWorkflowsQueueID(ctx, install.ID)
		if err != nil {
			ctx.Error(err)
			return
		}
		if err := s.enqueueInstallSignal(ctx, queueID, &executeflow.Signal{
			WorkflowID: workflow.ID,
		}, workflow.ID, "install_workflows"); err != nil {
			ctx.Error(fmt.Errorf("enqueue signal: %w", err))
			return
		}
	} else {
		s.evClient.Send(ctx, install.ID, &signals.Signal{
			Type:              signals.OperationExecuteFlow,
			InstallWorkflowID: workflow.ID,
		})
	}

	ctx.JSON(http.StatusCreated, app.WorkflowResponse{WorkflowID: workflow.ID})
}
