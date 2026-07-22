package service

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

var errDispatchNotRetryable = errors.New("dispatch is not retryable")

type eventResponse struct {
	app.EventSourceEvent
	Dispatches []app.EventDispatch `json:"dispatches"`
}

type replayResponse struct {
	EventID  string `json:"event_id"`
	ReplayID string `json:"replay_id"`
}

type retryResponse struct {
	DispatchID string `json:"dispatch_id"`
	RetryID    string `json:"retry_id"`
}

func listLimit(ctx *gin.Context) int {
	limit, err := strconv.Atoi(ctx.Query("limit"))
	if err != nil || limit <= 0 {
		return 50
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func (s *service) scopedEvent(ctx *gin.Context, orgID, eventID string) (*app.EventSourceEvent, error) {
	var event app.EventSourceEvent
	err := s.db.WithContext(ctx).Where(app.EventSourceEvent{ID: eventID, OrgID: orgID}).First(&event).Error
	return &event, err
}

func (s *service) scopedDispatch(ctx *gin.Context, orgID, dispatchID string) (*app.EventDispatch, error) {
	var dispatch app.EventDispatch
	err := s.db.WithContext(ctx).Where(app.EventDispatch{ID: dispatchID, OrgID: orgID}).First(&dispatch).Error
	return &dispatch, err
}

// @ID ListAutomationEvents
// @Summary List automation events
// @Tags event-automations
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param limit query int false "Maximum events to return (max 100)"
// @Success 200 {array} app.EventSourceEvent
// @Router /v1/event-automations/events [get]
func (s *service) ListEvents(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	var events []app.EventSourceEvent
	if err := s.db.WithContext(ctx).Where(app.EventSourceEvent{OrgID: org.ID}).Order("received_at DESC").Limit(listLimit(ctx)).Find(&events).Error; err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, events)
}

// @ID GetAutomationEvent
// @Summary Get an automation event and its dispatches
// @Tags event-automations
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param event_id path string true "Event ID"
// @Success 200 {object} eventResponse
// @Router /v1/event-automations/events/{event_id} [get]
func (s *service) GetEvent(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	event, err := s.scopedEvent(ctx, org.ID, ctx.Param("event_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	var dispatches []app.EventDispatch
	if err := s.db.WithContext(ctx).Where(app.EventDispatch{OrgID: org.ID, EventSourceEventID: event.ID}).Order("created_at DESC").Find(&dispatches).Error; err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, eventResponse{EventSourceEvent: *event, Dispatches: dispatches})
}

// @ID ReplayAutomationEvent
// @Summary Replay an automation event
// @Tags event-automations
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param event_id path string true "Event ID"
// @Success 202 {object} replayResponse
// @Router /v1/event-automations/events/{event_id}/replay [post]
func (s *service) ReplayEvent(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	event, err := s.scopedEvent(ctx, org.ID, ctx.Param("event_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	queue, err := s.orgAutomationQueue(ctx, org.ID)
	if err != nil {
		ctx.Error(err)
		return
	}
	replayID := uuid.NewString()
	dedupeKey := "automation-event:replay:" + event.ID + ":" + replayID
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&app.EventSourceEvent{}).
			Where(app.EventSourceEvent{ID: event.ID, OrgID: org.ID}).
			Updates(map[string]any{"routing_status": app.EventRoutingStatusAccepted, "routing_error": "", "routing_started_at": nil, "routing_completed_at": nil, "match_count": 0, "dispatch_count": 0})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		_, err := s.queueClient.EnqueueSignalInTransaction(ctx, tx, &queueclient.EnqueueSignalRequest{
			QueueID: queue.ID, Signal: signal.NewRaw(signal.SignalType("automation-event"), map[string]any{"event_id": event.ID, "replay_id": replayID}),
			OwnerID: event.ID, OwnerType: plugins.TableName(s.db, app.EventSourceEvent{}), DedupeKey: &dedupeKey,
		})
		return err
	})
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusAccepted, replayResponse{EventID: event.ID, ReplayID: replayID})
}

// @ID ListAutomationDispatches
// @Summary List automation dispatches
// @Tags event-automations
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param limit query int false "Maximum dispatches to return (max 100)"
// @Success 200 {array} app.EventDispatch
// @Router /v1/event-automations/dispatches [get]
func (s *service) ListDispatches(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	var dispatches []app.EventDispatch
	if err := s.db.WithContext(ctx).Where(app.EventDispatch{OrgID: org.ID}).Order("created_at DESC").Limit(listLimit(ctx)).Find(&dispatches).Error; err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, dispatches)
}

// @ID GetAutomationDispatch
// @Summary Get an automation dispatch
// @Tags event-automations
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param dispatch_id path string true "Dispatch ID"
// @Success 200 {object} app.EventDispatch
// @Router /v1/event-automations/dispatches/{dispatch_id} [get]
func (s *service) GetDispatch(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	dispatch, err := s.scopedDispatch(ctx, org.ID, ctx.Param("dispatch_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, dispatch)
}

// @ID RetryAutomationDispatch
// @Summary Retry an automation dispatch
// @Tags event-automations
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param dispatch_id path string true "Dispatch ID"
// @Success 202 {object} retryResponse
// @Failure 409 {object} map[string]string
// @Router /v1/event-automations/dispatches/{dispatch_id}/retry [post]
func (s *service) RetryDispatch(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	dispatch, err := s.scopedDispatch(ctx, org.ID, ctx.Param("dispatch_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	queue, err := s.appAutomationQueue(ctx, dispatch.AppID)
	if err != nil {
		ctx.Error(err)
		return
	}
	retryID := uuid.NewString()
	generationToken := uuid.NewString()
	dedupeKey := "automation-dispatch:retry:" + dispatch.ID + ":" + retryID
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked app.EventDispatch
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.EventDispatch{ID: dispatch.ID, OrgID: org.ID, AppID: dispatch.AppID}).First(&locked).Error; err != nil {
			return err
		}
		if locked.Status != app.EventDispatchStatusDeadLettered {
			return errDispatchNotRetryable
		}
		response, err := s.queueClient.EnqueueSignalInTransaction(ctx, tx, &queueclient.EnqueueSignalRequest{
			QueueID: queue.ID, Signal: signal.NewRaw(signal.SignalType("automation-dispatch"), map[string]any{"dispatch_id": dispatch.ID, "generation_token": generationToken}),
			OwnerID: dispatch.ID, OwnerType: plugins.TableName(s.db, app.EventDispatch{}), DedupeKey: &dedupeKey,
		})
		if err != nil {
			return err
		}
		return tx.Model(&locked).Updates(map[string]any{
			"status": app.EventDispatchStatusPending, "attempts": 0, "error": "", "next_attempt_at": nil,
			"generation_token": generationToken, "execution_token": "", "queue_signal_id": response.ID,
			"started_at": nil, "triggered_at": nil, "failed_at": nil,
		}).Error
	})
	if errors.Is(err, errDispatchNotRetryable) {
		ctx.JSON(http.StatusConflict, gin.H{"error": "dispatch is not retryable"})
		return
	}
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusAccepted, retryResponse{DispatchID: dispatch.ID, RetryID: retryID})
}
