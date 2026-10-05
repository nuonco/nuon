package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/request"
)

type ReprovisionInstallStackRequest struct {
	RequestID string `json:"request_id,omitempty" validate:"omitempty,max=255"`
	Role      string `json:"role,omitempty"`
	PlanOnly  bool   `json:"plan_only"`
}

// @ID						ReprovisionInstallStack
// @Summary				reprovision an install stack
// @Description.markdown	reprovision_install_stack.md
// @Param					install_id	path	string							true	"install ID"
// @Param					req			body	ReprovisionInstallStackRequest	true	"Input"
// @Tags					installs
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.WorkflowResponse
// @Router					/v1/installs/{install_id}/reprovision-stack [post]
func (s *service) ReprovisionInstallStack(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req ReprovisionInstallStackRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
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
			WorkflowType: app.WorkflowTypeReprovisionStack,
			Metadata:     map[string]string{},
			PlanOnly:     req.PlanOnly,
			Role:         req.Role,
			RequestID:    req.RequestID,
			RequestHash:  hash,
			Operation:    "reprovision-stack",
			QueueName:    installhelpers.InstallWorkflowsQueueName,
		}, nil)
		if err != nil {
			ctx.Error(err)
			return
		}
		s.logFlowAPIAction(ctx, "workflow.reprovision_stack_requested",
			zap.String("workflow_id", workflow.ID),
			zap.String("install_id", install.ID),
			zap.Bool("plan_only", req.PlanOnly),
		)
		ctx.JSON(http.StatusCreated, app.WorkflowResponse{WorkflowID: workflow.ID})
		return
	}

	workflow, err := s.helpers.CreateWorkflowWithRole(ctx,
		install.ID,
		app.WorkflowTypeReprovisionStack,
		map[string]string{},
		req.PlanOnly,
		req.Role,
	)
	if err != nil {
		ctx.Error(err)
		return
	}
	queueID, err := s.getInstallWorkflowsQueueID(ctx, install.ID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.enqueueInstallSignal(ctx, queueID, executeflow.NewSignal(workflow.ID), workflow.ID, "install_workflows"); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}

	s.logFlowAPIAction(ctx, "workflow.reprovision_stack_requested",
		zap.String("workflow_id", workflow.ID),
		zap.String("install_id", install.ID),
		zap.Bool("plan_only", req.PlanOnly),
	)

	ctx.JSON(http.StatusCreated, app.WorkflowResponse{WorkflowID: workflow.ID})
}
