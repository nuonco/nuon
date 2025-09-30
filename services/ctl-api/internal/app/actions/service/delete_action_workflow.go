package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/actions/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

// @ID						DeleteActionWorkflow
// @Summary				delete an action workflow
// @Description.markdown	delete_action_workflow.md
// @Param					action_workflow_id	path	string	true	"action workflow ID"
// @Tags					actions
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{boolean}	true
// @Router					/v1/action-workflows/{action_workflow_id} [DELETE]
func (s *service) DeleteActionWorkflow(ctx *gin.Context) {
	awID := ctx.Param("action_workflow_id")
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	aw, err := s.getActionWorkflowWithOrg(ctx, org.ID, awID)
	if err != nil {
		ctx.Error(err)
		return
	}

	err = s.deleteActionWorkflow(ctx, org.ID, aw.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	// trigger signal
	s.evClient.Send(ctx, aw.ID, &signals.Signal{
		Type: signals.OperationDelete,
	})

	ctx.JSON(http.StatusOK, true)
}

func (s *service) deleteActionWorkflow(ctx context.Context, orgID, awID string) error {
	aw := app.ActionWorkflow{
		ID:                awID,
		Status:            app.ActionWorkflowStatusDeleteQueued,
		StatusDescription: "Delete Queued",
	}

	resp := s.db.WithContext(ctx).
		Where("org_id = ? AND id = ?", orgID, awID).
		Updates(&aw)
	if resp.Error != nil {
		return fmt.Errorf("unable to delete action workflow %s: %w", awID, resp.Error)
	}

	if resp.RowsAffected == 0 {
		return fmt.Errorf("action workflow %s not found in org %s", awID, orgID)
	}

	return nil
}

func (s *service) getActionWorkflowWithOrg(ctx context.Context, orgID, awID string) (*app.ActionWorkflow, error) {
	actionWorkflow := app.ActionWorkflow{}
	res := s.db.WithContext(ctx).
		Where("org_id = ? AND (id = ? or name = ?)", orgID, awID, awID).
		First(&actionWorkflow)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get action workflow: %w", res.Error)
	}

	return &actionWorkflow, nil
}
