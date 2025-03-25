package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/signals"
)

type RestartActionWorkflowRequest struct{}

//	@ID						AdminRestartActionWorkflow
//	@Summary				restart an action workflow event loop
//	@Description.markdown	restart_action_workflow.md
//	@Param					action_workflow_id	path	string							true	"action ID"
//	@Param					req					body	RestartActionWorkflowRequest	true	"Input"
//	@Tags					actions/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Produce				json
//	@Success				200	{boolean}	true
//	@Router					/v1/action-workflows/{action_workflow_id}/admin-restart [POST]
func (s *service) RestartAction(ctx *gin.Context) {
	actionWorkflowID := ctx.Param("action_workflow_id")

	var req RestartActionWorkflowRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	fmt.Println(actionWorkflowID)

	actionWorkflow, err := s.getActionWorkflow(ctx, actionWorkflowID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install: %w", err))
		return
	}

	s.evClient.Send(ctx, actionWorkflow.ID, &signals.Signal{
		Type: signals.OperationRestart,
	})
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
