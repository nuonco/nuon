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

type SyncSecretsRequest struct {
	RequestID string `json:"request_id,omitempty" validate:"omitempty,max=255"`
	PlanOnly  bool   `json:"plan_only"`
}

// @ID						SyncSecrets
// @Summary				sync secrets install
// @Description.markdown sync_secrets.md
// @Param					install_id	path	string							true	"install ID"
// @Param					req			body	SyncSecretsRequest	false	"Input"
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
// @Router					/v1/installs/{install_id}/sync-secrets [post]
func (s *service) SyncSecrets(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req SyncSecretsRequest
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
			WorkflowType: app.WorkflowTypeSyncSecrets,
			Metadata:     map[string]string{},
			PlanOnly:     req.PlanOnly,
			Role:         "",
			RequestID:    req.RequestID,
			RequestHash:  hash,
			Operation:    "sync-secrets",
			QueueName:    installhelpers.InstallWorkflowsQueueName,
		}, nil)
		if err != nil {
			ctx.Error(err)
			return
		}
		ctx.JSON(http.StatusCreated, app.WorkflowResponse{WorkflowID: workflow.ID})
		return
	}

	workflow, err := s.helpers.CreateWorkflow(ctx,
		installID,
		app.WorkflowTypeSyncSecrets,
		map[string]string{},
		req.PlanOnly,
		nil,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	queueID, err := s.getInstallWorkflowsQueueID(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.enqueueInstallSignal(ctx, queueID, executeflow.NewSignal(workflow.ID), workflow.ID, "install_workflows"); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, app.WorkflowResponse{WorkflowID: workflow.ID})
}
