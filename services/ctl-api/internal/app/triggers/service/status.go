package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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

var (
	errDispatchNotRetryable = errors.New("dispatch is not retryable")
	errEventRejected        = errors.New("event is rejected")
	errEventNotReplayable   = errors.New("event routing is not complete")
)

type eventResponse struct {
	app.TriggerEvent
	Dispatches          []eventDispatchResponse `json:"dispatches"`
	DispatchesTruncated bool                    `json:"dispatches_truncated"`
	WaiterMatches       []eventWaiterMatch      `json:"waiter_matches"`
}

type eventDispatchResponse struct {
	app.EventDispatch
	InstallID   string `json:"install_id,omitempty"`
	RunbookID   string `json:"runbook_id,omitempty"`
	RunbookName string `json:"runbook_name,omitempty"`
}

type eventWaiterMatch struct {
	app.EventRunbookWaiter
	TriggerName      string `json:"trigger_name,omitempty"`
	MatchedEventType string `json:"matched_event_type,omitempty"`
	WorkflowStepName string `json:"workflow_step_name,omitempty"`
	RunbookRunID     string `json:"runbook_run_id,omitempty"`
	RunbookID        string `json:"runbook_id,omitempty"`
	RunbookName      string `json:"runbook_name,omitempty"`
}

type eventSummaryResponse struct {
	ID                  string                 `json:"id"`
	TriggerID           string                 `json:"trigger_id"`
	TriggerName         string                 `json:"trigger_name,omitempty"`
	ExternalID          string                 `json:"external_id"`
	Source              string                 `json:"source,omitempty"`
	EventType           string                 `json:"event_type"`
	OccurredAt          *time.Time             `json:"occurred_at,omitempty"`
	ReceivedAt          time.Time              `json:"received_at"`
	RoutingStatus       app.EventRoutingStatus `json:"routing_status"`
	RoutingError        string                 `json:"routing_error,omitempty"`
	RoutingStartedAt    *time.Time             `json:"routing_started_at,omitempty"`
	RoutingCompletedAt  *time.Time             `json:"routing_completed_at,omitempty"`
	MatchCount          int                    `json:"match_count"`
	WaiterMatchCount    int                    `json:"waiter_match_count"`
	DispatchCount       int                    `json:"dispatch_count"`
	Dispatches          []dispatchSummary      `json:"dispatches"`
	DispatchesTruncated bool                   `json:"dispatches_truncated"`
}

type dispatchSummary struct {
	ID     string                  `json:"id"`
	Status app.EventDispatchStatus `json:"status"`
	Error  string                  `json:"error,omitempty"`
}

type rankedDispatchSummary struct {
	ID                   string
	TriggerEventID       string
	Status               app.EventDispatchStatus
	Error                string
	CumulativeDispatches int `gorm:"column:cumulative_dispatches"`
}

type eventListResponse struct {
	Items      []eventSummaryResponse `json:"items"`
	NextCursor string                 `json:"next_cursor,omitempty"`
}

type eventListCursor struct {
	ReceivedAt time.Time `json:"received_at"`
	ID         string    `json:"id"`
}

type eventListFilters struct {
	Query          string
	EventType      string
	Outcome        string
	ReceivedAfter  *time.Time
	ReceivedBefore *time.Time
}

const embeddedDispatchLimit = 20

type dispatchListResponse struct {
	Items      []eventDispatchResponse `json:"items"`
	NextCursor string                  `json:"next_cursor,omitempty"`
}

type dispatchListCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        string    `json:"id"`
}

func encodeDispatchListCursor(dispatch app.EventDispatch) string {
	encoded, _ := json.Marshal(dispatchListCursor{CreatedAt: dispatch.CreatedAt, ID: dispatch.ID})
	return base64.RawURLEncoding.EncodeToString(encoded)
}

func decodeDispatchListCursor(value string) (*dispatchListCursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("decode cursor: %w", err)
	}
	var cursor dispatchListCursor
	decoder := json.NewDecoder(bytes.NewReader(decoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cursor); err != nil || cursor.CreatedAt.IsZero() || cursor.ID == "" {
		return nil, errors.New("invalid cursor")
	}
	return &cursor, nil
}

func encodeEventListCursor(event app.TriggerEvent) string {
	encoded, _ := json.Marshal(eventListCursor{ReceivedAt: event.ReceivedAt, ID: event.ID})
	return base64.RawURLEncoding.EncodeToString(encoded)
}

func decodeEventListCursor(value string) (*eventListCursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("decode cursor: %w", err)
	}
	var cursor eventListCursor
	decoder := json.NewDecoder(bytes.NewReader(decoded))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cursor); err != nil || cursor.ReceivedAt.IsZero() || cursor.ID == "" {
		return nil, errors.New("invalid cursor")
	}
	return &cursor, nil
}

type eventRawResponse struct {
	RawBodyBase64  string `json:"raw_body_base64"`
	RawBodySHA256  string `json:"raw_body_sha256"`
	RawBodySize    int64  `json:"raw_body_size"`
	RawContentType string `json:"raw_content_type,omitempty"`
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

func parseEventListFilters(ctx *gin.Context) (eventListFilters, error) {
	filters := eventListFilters{
		Query:     strings.TrimSpace(ctx.Query("query")),
		EventType: strings.TrimSpace(ctx.Query("event_type")),
		Outcome:   strings.TrimSpace(ctx.Query("outcome")),
	}

	if filters.Outcome != "" {
		switch filters.Outcome {
		case "ok", "ignored", "rejected", "processing", "failed":
		default:
			return eventListFilters{}, fmt.Errorf("invalid outcome %q", filters.Outcome)
		}
	}

	for value, destination := range map[string]**time.Time{
		"received_after":  &filters.ReceivedAfter,
		"received_before": &filters.ReceivedBefore,
	} {
		raw := strings.TrimSpace(ctx.Query(value))
		if raw == "" {
			continue
		}
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return eventListFilters{}, fmt.Errorf("invalid %s timestamp", value)
		}
		*destination = &parsed
	}
	if filters.ReceivedAfter != nil && filters.ReceivedBefore != nil && filters.ReceivedAfter.After(*filters.ReceivedBefore) {
		return eventListFilters{}, errors.New("received_after must not be later than received_before")
	}

	return filters, nil
}

func eventRoutingStatuses(outcome string) []app.EventRoutingStatus {
	switch outcome {
	case "ok":
		return []app.EventRoutingStatus{app.EventRoutingStatusMatched}
	case "ignored":
		return []app.EventRoutingStatus{app.EventRoutingStatusIgnored}
	case "rejected":
		return []app.EventRoutingStatus{app.EventRoutingStatusRejected}
	case "failed":
		return []app.EventRoutingStatus{app.EventRoutingStatusRoutingFailed}
	case "processing":
		return []app.EventRoutingStatus{app.EventRoutingStatusAccepted, app.EventRoutingStatusRouting}
	default:
		return nil
	}
}

func caseInsensitiveContains(column, value string) clause.Expression {
	return clause.Expr{
		SQL:  "LOWER(?) LIKE ?",
		Vars: []any{clause.Column{Name: column}, "%" + strings.ToLower(value) + "%"},
	}
}

func triggerScopedEventSearch(value string) clause.Expression {
	return clause.Or(
		caseInsensitiveContains("id", value),
		caseInsensitiveContains("external_id", value),
	)
}

func (s *service) scopedEvent(ctx *gin.Context, orgID, eventID string) (*app.TriggerEvent, error) {
	var event app.TriggerEvent
	err := s.db.WithContext(ctx).Omit("raw_body").Preload("Trigger", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).Where(app.TriggerEvent{ID: eventID, OrgID: orgID}).First(&event).Error
	event.TriggerName = event.Trigger.Name
	return &event, err
}

func (s *service) scopedEventRaw(ctx *gin.Context, orgID, eventID string) (*app.TriggerEvent, error) {
	var event app.TriggerEvent
	err := s.db.WithContext(ctx).
		Select("raw_body", "raw_body_sha256", "raw_body_size", "raw_content_type").
		Where(app.TriggerEvent{ID: eventID, OrgID: orgID}).
		First(&event).Error
	return &event, err
}

func (s *service) scopedDispatch(ctx *gin.Context, orgID, dispatchID string) (*app.EventDispatch, error) {
	var dispatch app.EventDispatch
	err := s.db.WithContext(ctx).Where(app.EventDispatch{ID: dispatchID, OrgID: orgID}).First(&dispatch).Error
	return &dispatch, err
}

func (s *service) listEvents(ctx *gin.Context, orgID, triggerID string, filters eventListFilters, cursor *eventListCursor, order string) {
	limit := listLimit(ctx)
	var events []app.TriggerEvent
	query := s.db.WithContext(ctx).
		Select("id", "trigger_id", "external_id", "source", "event_type", "occurred_at", "received_at", "routing_status", "routing_error", "routing_started_at", "routing_completed_at", "match_count", "waiter_match_count", "dispatch_count").
		Preload("Trigger", func(db *gorm.DB) *gorm.DB { return db.Unscoped() }).
		Where(app.TriggerEvent{OrgID: orgID, TriggerID: triggerID})
	if filters.EventType != "" {
		query = query.Where(app.TriggerEvent{EventType: filters.EventType})
	}
	if statuses := eventRoutingStatuses(filters.Outcome); len(statuses) == 1 {
		query = query.Where(app.TriggerEvent{RoutingStatus: statuses[0]})
	} else if len(statuses) > 1 {
		query = query.Where(map[string]any{"routing_status": statuses})
	}
	if filters.ReceivedAfter != nil {
		query = query.Where(clause.Gte{Column: clause.Column{Name: "received_at"}, Value: *filters.ReceivedAfter})
	}
	if filters.ReceivedBefore != nil {
		query = query.Where(clause.Lte{Column: clause.Column{Name: "received_at"}, Value: *filters.ReceivedBefore})
	}
	if filters.Query != "" {
		query = query.Where(triggerScopedEventSearch(filters.Query))
	}
	if cursor != nil {
		query = query.Where(eventCursorExpression(cursor, order))
	}
	orderClause := "received_at DESC, id DESC"
	if order == "asc" {
		orderClause = "received_at ASC, id ASC"
	}
	if err := query.
		Order(orderClause).
		Limit(limit + 1).
		Find(&events).Error; err != nil {
		ctx.Error(err)
		return
	}
	nextCursor := ""
	if len(events) > limit {
		events = events[:limit]
		nextCursor = encodeEventListCursor(events[len(events)-1])
	}
	eventIDs := make([]string, len(events))
	for i := range events {
		eventIDs[i] = events[i].ID
	}
	var dispatches []rankedDispatchSummary
	if len(eventIDs) > 0 {
		rankedDispatches := s.db.WithContext(ctx).
			Model(&app.EventDispatch{}).
			Select("id", "trigger_event_id", "status", "error", "created_at", "ROW_NUMBER() OVER (PARTITION BY trigger_event_id ORDER BY created_at DESC, id DESC) AS dispatch_rank", "COUNT(*) OVER (PARTITION BY trigger_event_id) AS cumulative_dispatches").
			Where(app.EventDispatch{OrgID: orgID}).
			Where(map[string]any{"trigger_event_id": eventIDs})
		if err := s.db.WithContext(ctx).
			Table("(?) AS ranked_dispatches", rankedDispatches).
			Where(clause.Lte{Column: clause.Column{Name: "dispatch_rank"}, Value: embeddedDispatchLimit}).
			Find(&dispatches).Error; err != nil {
			ctx.Error(err)
			return
		}
	}
	dispatchesByEvent := make(map[string][]dispatchSummary)
	cumulativeDispatchesByEvent := make(map[string]int)
	for _, dispatch := range dispatches {
		cumulativeDispatchesByEvent[dispatch.TriggerEventID] = dispatch.CumulativeDispatches
		if len(dispatchesByEvent[dispatch.TriggerEventID]) >= embeddedDispatchLimit {
			continue
		}
		dispatchesByEvent[dispatch.TriggerEventID] = append(dispatchesByEvent[dispatch.TriggerEventID], dispatchSummary{ID: dispatch.ID, Status: dispatch.Status, Error: dispatch.Error})
	}
	summaries := make([]eventSummaryResponse, len(events))
	for i := range events {
		event := &events[i]
		summaries[i] = eventSummaryResponse{
			ID: event.ID, TriggerID: event.TriggerID, TriggerName: event.Trigger.Name,
			ExternalID: event.ExternalID, Source: event.Source, EventType: event.EventType, OccurredAt: event.OccurredAt,
			ReceivedAt: event.ReceivedAt, RoutingStatus: event.RoutingStatus, RoutingError: event.RoutingError,
			RoutingStartedAt: event.RoutingStartedAt, RoutingCompletedAt: event.RoutingCompletedAt,
			MatchCount: event.MatchCount, WaiterMatchCount: event.WaiterMatchCount, DispatchCount: event.DispatchCount,
			Dispatches:          dispatchesByEvent[event.ID],
			DispatchesTruncated: cumulativeDispatchesByEvent[event.ID] > len(dispatchesByEvent[event.ID]),
		}
	}
	ctx.JSON(http.StatusOK, eventListResponse{Items: summaries, NextCursor: nextCursor})
}

func eventCursorExpression(cursor *eventListCursor, order string) clause.Expression {
	operator := "<"
	if order == "asc" {
		operator = ">"
	}
	return clause.Expr{SQL: "received_at " + operator + " ? OR (received_at = ? AND id " + operator + " ?)", Vars: []any{cursor.ReceivedAt, cursor.ReceivedAt, cursor.ID}}
}

// @ID GetTriggerEvent
// @Summary Get a trigger event and its dispatches
// @Tags triggers
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param event_id path string true "Event ID"
// @Success 200 {object} eventResponse
// @Router /v1/triggers/events/{event_id} [get]
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
	if err := s.db.WithContext(ctx).Where(app.EventDispatch{OrgID: org.ID, TriggerEventID: event.ID}).Order("created_at DESC, id DESC").Limit(embeddedDispatchLimit + 1).Find(&dispatches).Error; err != nil {
		ctx.Error(err)
		return
	}
	truncated := len(dispatches) > embeddedDispatchLimit
	if truncated {
		dispatches = dispatches[:embeddedDispatchLimit]
	}
	dispatchResponses, err := s.eventDispatchResponses(ctx, dispatches)
	if err != nil {
		ctx.Error(err)
		return
	}
	waiterMatches, err := s.eventWaiterMatches(ctx, org.ID, event)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, eventResponse{
		TriggerEvent:        *event,
		Dispatches:          dispatchResponses,
		DispatchesTruncated: truncated,
		WaiterMatches:       waiterMatches,
	})
}

func (s *service) eventDispatchResponses(ctx *gin.Context, dispatches []app.EventDispatch) ([]eventDispatchResponse, error) {
	responses := make([]eventDispatchResponse, len(dispatches))
	for i := range dispatches {
		dispatch := dispatches[i]
		response := eventDispatchResponse{EventDispatch: dispatch}
		if dispatch.ResultResourceType != "install_runbook_runs" || dispatch.ResultResourceID == "" {
			responses[i] = response
			continue
		}
		var run app.InstallRunbookRun
		if err := s.db.WithContext(ctx).Preload("InstallRunbook.Runbook").Where(app.InstallRunbookRun{ID: dispatch.ResultResourceID, OrgID: dispatch.OrgID}).First(&run).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				responses[i] = response
				continue
			}
			return nil, err
		}
		response.InstallID = run.InstallID
		response.RunbookID = run.InstallRunbook.RunbookID
		response.RunbookName = run.InstallRunbook.Runbook.Name
		if run.InstallWorkflowID != nil {
			response.WorkflowID = *run.InstallWorkflowID
		}
		responses[i] = response
	}
	return responses, nil
}

func (s *service) eventWaiterMatches(ctx *gin.Context, orgID string, event *app.TriggerEvent) ([]eventWaiterMatch, error) {
	var waiters []app.EventRunbookWaiter
	if err := s.db.WithContext(ctx).Where(app.EventRunbookWaiter{OrgID: orgID, MatchedEventID: &event.ID}).Order("matched_at DESC, id DESC").Find(&waiters).Error; err != nil {
		return nil, err
	}
	matches := make([]eventWaiterMatch, len(waiters))
	for i := range waiters {
		waiter := waiters[i]
		match := eventWaiterMatch{EventRunbookWaiter: waiter, TriggerName: event.TriggerName, MatchedEventType: event.EventType}
		var step app.WorkflowStep
		if err := s.db.WithContext(ctx).Select("name").Where(app.WorkflowStep{ID: waiter.WorkflowStepID, OrgID: orgID}).First(&step).Error; err == nil {
			match.WorkflowStepName = step.Name
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		var run app.InstallRunbookRun
		if err := s.db.WithContext(ctx).Preload("InstallRunbook.Runbook").Where(app.InstallRunbookRun{OrgID: orgID, InstallWorkflowID: &waiter.WorkflowID}).First(&run).Error; err == nil {
			match.RunbookRunID = run.ID
			match.RunbookID = run.InstallRunbook.RunbookID
			match.RunbookName = run.InstallRunbook.Runbook.Name
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		matches[i] = match
	}
	return matches, nil
}

// @ID GetTriggerEventRaw
// @Summary Get the raw request body for a trigger event
// @Tags triggers
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param event_id path string true "Event ID"
// @Success 200 {object} eventRawResponse
// @Router /v1/triggers/events/{event_id}/raw [get]
func (s *service) GetEventRaw(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	event, err := s.scopedEventRaw(ctx, org.ID, ctx.Param("event_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, eventRawResponse{
		RawBodyBase64:  base64.StdEncoding.EncodeToString(event.RawBody),
		RawBodySHA256:  event.RawBodySHA256,
		RawBodySize:    event.RawBodySize,
		RawContentType: event.RawContentType,
	})
}

// @ID ReplayTriggerEvent
// @Summary Replay a trigger event
// @Tags triggers
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param event_id path string true "Event ID"
// @Success 202 {object} replayResponse
// @Router /v1/triggers/events/{event_id}/replay [post]
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
	if event.RoutingStatus == app.EventRoutingStatusRejected {
		ctx.JSON(http.StatusConflict, gin.H{"error": "rejected events cannot be replayed"})
		return
	}
	if !eventReplayable(event.RoutingStatus) {
		ctx.JSON(http.StatusConflict, gin.H{"error": "event can only be replayed after routing completes"})
		return
	}
	queue, err := s.orgTriggerQueue(ctx, org.ID)
	if err != nil {
		ctx.Error(err)
		return
	}
	replayID := uuid.NewString()
	dedupeKey := "trigger-event:replay:" + event.ID + ":" + replayID
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked app.TriggerEvent
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.TriggerEvent{ID: event.ID, OrgID: org.ID}).First(&locked).Error; err != nil {
			return err
		}
		if locked.RoutingStatus == app.EventRoutingStatusRejected {
			return errEventRejected
		}
		if !eventReplayable(locked.RoutingStatus) {
			return errEventNotReplayable
		}
		_, err := s.queueClient.EnqueueSignalInTransaction(ctx, tx, &queueclient.EnqueueSignalRequest{
			QueueID: queue.ID, Signal: signal.NewRaw(signal.SignalType("trigger-event"), map[string]any{"event_id": event.ID, "replay_id": replayID}),
			OwnerID: event.ID, OwnerType: plugins.TableName(s.db, app.TriggerEvent{}), DedupeKey: &dedupeKey,
		})
		return err
	})
	if errors.Is(err, errEventRejected) {
		ctx.JSON(http.StatusConflict, gin.H{"error": "rejected events cannot be replayed"})
		return
	}
	if errors.Is(err, errEventNotReplayable) {
		ctx.JSON(http.StatusConflict, gin.H{"error": "event can only be replayed after routing completes"})
		return
	}
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusAccepted, replayResponse{EventID: event.ID, ReplayID: replayID})
}

func eventReplayable(status app.EventRoutingStatus) bool {
	return status == app.EventRoutingStatusMatched || status == app.EventRoutingStatusIgnored || status == app.EventRoutingStatusRoutingFailed
}

// @ID ListTriggerEventDispatches
// @Summary List trigger event dispatches
// @Tags triggers
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param limit query int false "Maximum dispatches to return (max 100)"
// @Param event_id query string false "Event ID"
// @Param cursor query string false "Opaque pagination cursor"
// @Success 200 {object} dispatchListResponse
// @Router /v1/triggers/dispatches [get]
func (s *service) ListDispatches(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	var cursor *dispatchListCursor
	if value := ctx.Query("cursor"); value != "" {
		cursor, err = decodeDispatchListCursor(value)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid cursor"})
			return
		}
	}
	limit := listLimit(ctx)
	var dispatches []app.EventDispatch
	query := s.db.WithContext(ctx).Where(app.EventDispatch{OrgID: org.ID})
	if eventID := ctx.Query("event_id"); eventID != "" {
		query = query.Where(app.EventDispatch{TriggerEventID: eventID})
	}
	if cursor != nil {
		query = query.Where(clause.Or(
			clause.Lt{Column: clause.Column{Name: "created_at"}, Value: cursor.CreatedAt},
			clause.And(
				clause.Eq{Column: clause.Column{Name: "created_at"}, Value: cursor.CreatedAt},
				clause.Lt{Column: clause.Column{Name: "id"}, Value: cursor.ID},
			),
		))
	}
	if err := query.Order("created_at DESC, id DESC").Limit(limit + 1).Find(&dispatches).Error; err != nil {
		ctx.Error(err)
		return
	}
	nextCursor := ""
	if len(dispatches) > limit {
		dispatches = dispatches[:limit]
		nextCursor = encodeDispatchListCursor(dispatches[len(dispatches)-1])
	}
	responses, err := s.eventDispatchResponses(ctx, dispatches)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, dispatchListResponse{Items: responses, NextCursor: nextCursor})
}

// @ID GetTriggerEventDispatch
// @Summary Get a trigger event dispatch
// @Tags triggers
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param dispatch_id path string true "Dispatch ID"
// @Success 200 {object} app.EventDispatch
// @Router /v1/triggers/dispatches/{dispatch_id} [get]
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
	responses, err := s.eventDispatchResponses(ctx, []app.EventDispatch{*dispatch})
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, responses[0])
}

// @ID RetryTriggerEventDispatch
// @Summary Retry a trigger event dispatch
// @Tags triggers
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param dispatch_id path string true "Dispatch ID"
// @Success 202 {object} retryResponse
// @Failure 409 {object} map[string]string
// @Router /v1/triggers/dispatches/{dispatch_id}/retry [post]
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
	queue, err := s.appTriggerQueue(ctx, dispatch.AppID)
	if err != nil {
		ctx.Error(err)
		return
	}
	retryID := uuid.NewString()
	generationToken := uuid.NewString()
	dedupeKey := "trigger-event-dispatch:retry:" + dispatch.ID + ":" + retryID
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked app.EventDispatch
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.EventDispatch{ID: dispatch.ID, OrgID: org.ID, AppID: dispatch.AppID}).First(&locked).Error; err != nil {
			return err
		}
		if locked.Status != app.EventDispatchStatusDeadLettered {
			return errDispatchNotRetryable
		}
		response, err := s.queueClient.EnqueueSignalInTransaction(ctx, tx, &queueclient.EnqueueSignalRequest{
			QueueID: queue.ID, Signal: signal.NewRaw(signal.SignalType("trigger-event-dispatch"), map[string]any{"dispatch_id": dispatch.ID, "generation_token": generationToken}),
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
