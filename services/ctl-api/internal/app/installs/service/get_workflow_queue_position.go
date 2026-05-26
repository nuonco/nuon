package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// WorkflowQueuePositionResponse describes the queue position of a workflow.
type WorkflowQueuePositionResponse struct {
	// Position is the 1-based queue position (1 = next to execute).
	Position int `json:"position"`
	// QueueDepth is the total number of signals waiting in the queue.
	QueueDepth int `json:"queue_depth"`
	// SignalsAhead are the workflows ahead in the queue, ordered from front to back.
	SignalsAhead []WorkflowQueueItem `json:"signals_ahead"`
}

// WorkflowQueueItem is a summary of a workflow waiting in the queue.
type WorkflowQueueItem struct {
	WorkflowID   string            `json:"workflow_id"`
	WorkflowType app.WorkflowType  `json:"workflow_type"`
	Status       app.Status        `json:"status"`
	CreatedAt    string            `json:"created_at"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// @ID						GetWorkflowQueuePosition
// @Summary					get queue position for a workflow
// @Description				Returns the queue position and workflows ahead when a workflow is pending.
// @Param					workflow_id	path	string	true	"workflow ID"
// @Tags					installs
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{object}	WorkflowQueuePositionResponse
// @Router					/v1/workflows/{workflow_id}/queue-position [GET]
func (s *service) GetWorkflowQueuePosition(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get org from context"))
		return
	}

	workflowID := ctx.Param("workflow_id")

	// Find the queue signal for this workflow.
	var qs app.QueueSignal
	res := s.db.WithContext(ctx).
		Where(app.QueueSignal{
			OwnerID:   workflowID,
			OwnerType: "install_workflows",
		}).
		Order("created_at DESC").
		First(&qs)
	if res.Error != nil {
		// Also check action workflows owner type.
		res = s.db.WithContext(ctx).
			Where(app.QueueSignal{
				OwnerID:   workflowID,
				OwnerType: "install_action_workflows",
			}).
			Order("created_at DESC").
			First(&qs)
		if res.Error != nil {
			ctx.JSON(http.StatusOK, WorkflowQueuePositionResponse{})
			return
		}
	}

	// Find signals ahead in the same queue (created before this one, not yet completed).
	var signalsAhead []app.QueueSignal
	s.db.WithContext(ctx).
		Where("queue_id = ? AND created_at < ? AND id != ?", qs.QueueID, qs.CreatedAt, qs.ID).
		Order("created_at ASC").
		Limit(20).
		Find(&signalsAhead)

	// Look up the workflows for the signals ahead.
	items := make([]WorkflowQueueItem, 0, len(signalsAhead))
	for _, sig := range signalsAhead {
		var wf app.Workflow
		wfRes := s.db.WithContext(ctx).
			Where("id = ? AND org_id = ?", sig.OwnerID, org.ID).
			First(&wf)
		if wfRes.Error != nil {
			continue
		}

		// Only include non-terminal workflows.
		if wf.Status.Status == app.StatusSuccess || wf.Status.Status == app.StatusError ||
			wf.Status.Status == app.StatusCancelled {
			continue
		}

		items = append(items, WorkflowQueueItem{
			WorkflowID:   wf.ID,
			WorkflowType: wf.Type,
			Status:       wf.Status.Status,
			CreatedAt:    wf.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	// Count total queue depth from non-terminal signals.
	var totalDepth int64
	s.db.WithContext(ctx).Model(&app.QueueSignal{}).
		Where("queue_id = ?", qs.QueueID).
		Count(&totalDepth)

	ctx.JSON(http.StatusOK, WorkflowQueuePositionResponse{
		Position:     len(items) + 1, // 1-based: this workflow is after the items ahead
		QueueDepth:   int(totalDepth),
		SignalsAhead: items,
	})
}
