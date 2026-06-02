package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type RestartActionWorkflowRequest struct{}

// @ID						AdminRestartActionWorkflow
// @Summary				restart an action workflow
// @Description.markdown	restart_action_workflow.md
// @Param					action_workflow_id	path	string							true	"action ID"
// @Param					req					body	RestartActionWorkflowRequest	true	"Input"
// @Tags					actions/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{boolean}	true
// @Router					/v1/action-workflows/{action_workflow_id}/admin-restart [POST]
func (s *service) RestartAction(ctx *gin.Context) {
	actionWorkflowID := ctx.Param("action_workflow_id")

	var req RestartActionWorkflowRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	_, err := s.getActionWorkflow(ctx, actionWorkflowID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) getActionWorkflow(ctx context.Context, installID string) (*app.ActionWorkflow, error) {
	actionWorkflow := app.ActionWorkflow{}
	res := s.db.WithContext(ctx).
		First(&actionWorkflow, "id = ?", installID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get action workflow: %w", res.Error)
	}

	return &actionWorkflow, nil
}
