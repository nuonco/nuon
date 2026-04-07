package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type ReprovisionInstallSandboxRequest struct {
	Role           string `json:"role,omitempty"`
	PlanOnly       bool   `json:"plan_only"`
	SkipComponents bool   `json:"skip_components"`
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
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	metadata := map[string]string{}
	if req.SkipComponents {
		metadata["skip_components"] = "true"
	}

	workflow, err := s.helpers.CreateWorkflowWithRole(ctx,
		install.ID,
		app.WorkflowTypeReprovisionSandbox,
		metadata,
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
			InstallWorkflowID: workflow.ID,
		}); err != nil {
			ctx.Error(fmt.Errorf("enqueue signal: %w", err))
			return
		}
	} else {
		s.evClient.Send(ctx, install.ID, &signals.Signal{
			Type:              signals.OperationExecuteFlow,
			InstallWorkflowID: workflow.ID,
		})
	}

	ctx.Header(app.HeaderInstallWorkflowID, workflow.ID)

	ctx.JSON(http.StatusCreated, "ok")
}
