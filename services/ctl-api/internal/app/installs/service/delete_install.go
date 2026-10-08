package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/lifecyclephase"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	forgotten "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/forgotten"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/request"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

// DEPRECATED: This endpoint is deprecated and will be removed in a future release.

type DeleteInstallRequest struct {
	RequestID string `json:"request_id,omitempty" validate:"omitempty,max=255"`
}

func (c *DeleteInstallRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						DeleteInstall
// @Summary				delete an install
// @Description.markdown	delete_install.md
// @Param					install_id	path	string	true	"install ID"
// @Param					req			body	DeleteInstallRequest	false	"Input"
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
// @Success				200	{object}	app.WorkflowResponse
// @Router					/v1/installs/{install_id} [DELETE]
func (s *service) DeleteInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req DeleteInstallRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
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
		workflow, created, err := s.helpers.RunIdempotentInstallWorkflow(ctx, installhelpers.IdempotentInstallWorkflowRequest{
			InstallID:    install.ID,
			WorkflowType: app.WorkflowTypeDeprovision,
			Metadata:     map[string]string{},
			RequestID:    req.RequestID,
			RequestHash:  hash,
			Operation:    "delete-install",
			QueueName:    installhelpers.InstallWorkflowsQueueName,
		}, nil)
		if err != nil {
			ctx.Error(err)
			return
		}
		if created {
			lp := lifecyclephase.New(lifecyclephase.Deprovisioning, "Tearing down components and cloud resources")
			s.db.WithContext(ctx).Model(&app.Install{ID: install.ID}).Updates(map[string]any{
				"lifecycle_phase": lp,
			})

			signalsQueueID, err := s.getInstallSignalsQueueID(ctx, install.ID)
			if err != nil {
				ctx.Error(err)
				return
			}
			if err := s.enqueueInstallSignal(ctx, signalsQueueID, &forgotten.Signal{
				InstallID: install.ID,
			}, "", ""); err != nil {
				ctx.Error(fmt.Errorf("enqueue signal: %w", err))
				return
			}
		}
		ctx.JSON(http.StatusOK, app.WorkflowResponse{WorkflowID: workflow.ID})
		return
	}

	workflow, err := s.helpers.CreateWorkflow(ctx,
		install.ID,
		app.WorkflowTypeDeprovision,
		map[string]string{},
		false,
		nil,
	)
	if err != nil {
		ctx.Error(err)
		return
	}

	lp := lifecyclephase.New(lifecyclephase.Deprovisioning, "Tearing down components and cloud resources")
	s.db.WithContext(ctx).Model(&app.Install{ID: install.ID}).Updates(map[string]any{
		"lifecycle_phase": lp,
	})

	workflowsQueueID, err := s.getInstallWorkflowsQueueID(ctx, install.ID)
	if err != nil {
		ctx.Error(err)
		return
	}
	signalsQueueID, err := s.getInstallSignalsQueueID(ctx, install.ID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.enqueueInstallSignal(ctx, workflowsQueueID, executeflow.NewSignal(workflow.ID), workflow.ID, "install_workflows"); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}
	if err := s.enqueueInstallSignal(ctx, signalsQueueID, &forgotten.Signal{
		InstallID: install.ID,
	}, "", ""); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, app.WorkflowResponse{WorkflowID: workflow.ID})
}
