package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CancelWorkflowsRequest struct {
	WorkflowIDs []string `json:"workflow_ids" binding:"required"`
}

type CancelWorkflowsResponse struct {
	Cancelled []string              `json:"cancelled"`
	Errors    []CancelWorkflowError `json:"errors,omitempty"`
}

type CancelWorkflowError struct {
	WorkflowID string `json:"workflow_id"`
	Error      string `json:"error"`
}

// @ID							CancelWorkflows
// @Summary						cancel multiple workflows
// @Description					Cancel multiple workflows by ID. Returns partial results if some fail.
// @Param						body	body	CancelWorkflowsRequest	true	"workflow IDs to cancel"
// @Tags						installs
// @Accept						json
// @Produce						json
// @Security					APIKey
// @Security					OrgID
// @Failure						400	{object}	stderr.ErrResponse
// @Failure						401	{object}	stderr.ErrResponse
// @Failure						403	{object}	stderr.ErrResponse
// @Failure						500	{object}	stderr.ErrResponse
// @Success				200	{object}	CancelWorkflowsResponse
// @Router						/v1/workflows/cancel [post]
func (s *service) CancelWorkflows(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org from context: %w", err))
		return
	}

	var req CancelWorkflowsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request body: %w", err))
		return
	}

	resp := CancelWorkflowsResponse{}

	var allowed map[string]bool
	if authz.HasAuthority(ctx) {
		ids, err := require.CollectionIDs(ctx)
		if err != nil {
			ctx.Error(err)
			return
		}
		var workflowIDs []string
		if err := s.db.WithContext(ctx.Request.Context()).Model(&app.Workflow{}).
			Where(&app.Workflow{OrgID: org.ID}).Where("id IN (?)", ids).
			Where(map[string]any{"id": req.WorkflowIDs}).Pluck("id", &workflowIDs).Error; err != nil {
			ctx.Error(err)
			return
		}
		allowed = make(map[string]bool, len(workflowIDs))
		for _, id := range workflowIDs {
			allowed[id] = true
		}
	}

	for _, workflowID := range req.WorkflowIDs {
		if allowed != nil && !allowed[workflowID] {
			resp.Errors = append(resp.Errors, CancelWorkflowError{WorkflowID: workflowID, Error: "workflow not found"})
			continue
		}
		if err := s.cancelSingleWorkflow(ctx, org.ID, workflowID); err != nil {
			resp.Errors = append(resp.Errors, CancelWorkflowError{
				WorkflowID: workflowID,
				Error:      err.Error(),
			})
		} else {
			resp.Cancelled = append(resp.Cancelled, workflowID)
		}
	}

	ctx.JSON(http.StatusOK, resp)
}
