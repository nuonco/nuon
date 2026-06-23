package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"

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

	// A signal still occupies the queue until its own status reaches a terminal
	// state. Workflow status is not a reliable proxy — a signal can be parked
	// (StatusPending, e.g. signal type not yet registered) or not-yet-enqueued
	// while its workflow is still non-terminal, and vice versa.
	terminalStatuses := []string{
		string(app.StatusSuccess),
		string(app.StatusError),
		string(app.StatusCancelled),
	}
	workflowOwnerTypes := []string{"install_workflows", "install_action_workflows"}

	inQueue := func(db *gorm.DB) *gorm.DB {
		return db.
			Where("queue_id = ?", qs.QueueID).
			Where("owner_type IN ?", workflowOwnerTypes).
			Where("status->>'status' NOT IN ?", terminalStatuses)
	}

	// Total workflows still waiting in this queue (includes this one).
	var queueDepth int64
	inQueue(s.db.WithContext(ctx).Model(&app.QueueSignal{})).Count(&queueDepth)

	// Workflows ahead of this one — those enqueued before it that are still in the queue.
	var aheadCount int64
	inQueue(s.db.WithContext(ctx).Model(&app.QueueSignal{})).
		Where("created_at < ? AND id != ?", qs.CreatedAt, qs.ID).
		Count(&aheadCount)

	// Fetch the signals immediately ahead for display, front to back.
	var signalsAhead []app.QueueSignal
	inQueue(s.db.WithContext(ctx)).
		Where("created_at < ? AND id != ?", qs.CreatedAt, qs.ID).
		Order("created_at ASC").
		Limit(20).
		Find(&signalsAhead)

	// Batch-load the workflows for the signals ahead to enrich the display items.
	ownerIDs := make([]string, 0, len(signalsAhead))
	for _, sig := range signalsAhead {
		ownerIDs = append(ownerIDs, sig.OwnerID)
	}

	workflowsByID := make(map[string]app.Workflow, len(ownerIDs))
	if len(ownerIDs) > 0 {
		var wfs []app.Workflow
		s.db.WithContext(ctx).
			Where("id IN ? AND org_id = ?", ownerIDs, org.ID).
			Find(&wfs)
		for _, wf := range wfs {
			workflowsByID[wf.ID] = wf
		}
	}

	items := make([]WorkflowQueueItem, 0, len(signalsAhead))
	for _, sig := range signalsAhead {
		item := WorkflowQueueItem{
			WorkflowID: sig.OwnerID,
			Status:     sig.Status.Status,
			CreatedAt:  sig.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		// Action workflows live in a different table and won't resolve here; they
		// still count toward position but display without an enriched type.
		if wf, ok := workflowsByID[sig.OwnerID]; ok {
			item.WorkflowType = wf.Type
			item.Status = wf.Status.Status
			item.CreatedAt = wf.CreatedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		items = append(items, item)
	}

	ctx.JSON(http.StatusOK, WorkflowQueuePositionResponse{
		Position:     int(aheadCount) + 1, // 1-based: this workflow is after the ones ahead
		QueueDepth:   int(queueDepth),
		SignalsAhead: items,
	})
}
