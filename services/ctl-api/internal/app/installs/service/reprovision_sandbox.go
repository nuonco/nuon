package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/request"
)

type ReprovisionInstallSandboxRequest struct {
	// RequestID is an optional idempotency key. The same id and body returns the original workflow. A different body, or an install that has moved to another app config, returns 409.
	RequestID      string `json:"request_id,omitempty" validate:"omitempty,max=255"`
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
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.WorkflowResponse
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

	if req.RequestID != "" {
		hashReq := req
		hashReq.RequestID = ""
		hash, err := request.Hash(hashReq)
		if err != nil {
			ctx.Error(err)
			return
		}
		workflow, _, err := s.helpers.RunIdempotentInstallWorkflow(ctx, installhelpers.IdempotentInstallWorkflowRequest{
			InstallID:    install.ID,
			WorkflowType: app.WorkflowTypeReprovisionSandbox,
			Metadata:     metadata,
			PlanOnly:     req.PlanOnly,
			Role:         req.Role,
			RequestID:    req.RequestID,
			RequestHash:  hash,
			Operation:    "reprovision-sandbox",
			QueueName:    installhelpers.InstallWorkflowsQueueName,
		}, nil)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.JSON(http.StatusCreated, app.WorkflowResponse{WorkflowID: workflow.ID})
		return
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
	queueID, err := s.getInstallWorkflowsQueueID(ctx, install.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("error queuing workflow %s: %w", workflow.ID, err))
		return
	}
	if err := s.enqueueInstallSignal(ctx, queueID, executeflow.NewSignal(workflow.ID), workflow.ID, "install_workflows"); err != nil {
		ctx.Error(fmt.Errorf("error queuing workflow %s: %w", workflow.ID, err))
		return
	}

	ctx.JSON(http.StatusCreated, app.WorkflowResponse{WorkflowID: workflow.ID})
}
