package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type AdminMarkOldQueueSignalsRequest struct {
	// Go duration, e.g. "24h", "72h", "30m".
	OlderThan    string     `json:"older_than" binding:"required"`
	Status       app.Status `json:"status" binding:"required"`
	ReasonFilter string     `json:"reason_filter"`
	Reason       string     `json:"reason"`
	OrgID        string     `json:"org_id"`
}

type AdminMarkOldQueueSignalsResponse struct {
	Changed int `json:"changed"`
}

// @ID						AdminMarkOldQueueSignals
// @Summary				Mark old queue signals
// @Description			Updates non-deleted queue signals that are currently queued or in-progress and older than
// @Description			older_than to the given status. Optionally filter by existing status_human_description via
// @Description			reason_filter, and stamp a new reason on the updated status. Scope with org_id when set.
// @Param					req	body	AdminMarkOldQueueSignalsRequest	true	"older_than, status, and optional filters"
// @Tags					queues/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{object}	AdminMarkOldQueueSignalsResponse
// @Failure				400	{object}	stderr.ErrResponse
// @Router					/v1/queues/admin-mark-old-signals [POST]
func (s *service) AdminMarkOldQueueSignals(ctx *gin.Context) {
	var req AdminMarkOldQueueSignalsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	olderThan, err := time.ParseDuration(req.OlderThan)
	if err != nil {
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("invalid older_than duration: %w", err)))
		return
	}
	if olderThan <= 0 {
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("older_than must be positive")))
		return
	}
	if req.Status == "" {
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("status is required")))
		return
	}

	cutoff := time.Now().Add(-olderThan)
	targetStatus := app.NewCompositeStatus(ctx, req.Status)
	if req.Reason != "" {
		targetStatus.StatusHumanDescription = req.Reason
	}

	var changed []app.QueueSignal
	query := s.db.WithContext(ctx).
		Model(&changed).
		Clauses(clause.Returning{}).
		Where("deleted_at = 0").
		Where("created_at < ?", cutoff).
		Where("status->>'status' IN ?", []string{
			string(app.StatusQueued),
			string(app.StatusInProgress),
		}).
		Where("status->>'status' != ?", req.Status)

	if req.ReasonFilter != "" {
		query = query.Where("status->>'status_human_description' = ?", req.ReasonFilter)
	}
	if req.OrgID != "" {
		query = query.Where("org_id = ?", req.OrgID)
	}

	if res := query.Updates(map[string]any{
		"status": targetStatus,
	}); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to mark old queue signals: %w", res.Error))
		return
	}

	s.l.Info("marked old queue signals",
		zap.String("older-than", req.OlderThan),
		zap.Time("cutoff", cutoff),
		zap.String("status", string(req.Status)),
		zap.String("reason-filter", req.ReasonFilter),
		zap.String("reason", req.Reason),
		zap.String("org-id", req.OrgID),
		zap.Int("changed", len(changed)),
	)

	ctx.JSON(http.StatusOK, AdminMarkOldQueueSignalsResponse{
		Changed: len(changed),
	})
}
