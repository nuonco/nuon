package service

import (
	"encoding/base64"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestListLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := map[string]struct {
		query string
		want  int
	}{
		"default":     {query: "", want: 50},
		"invalid":     {query: "?limit=nope", want: 50},
		"nonpositive": {query: "?limit=0", want: 50},
		"requested":   {query: "?limit=25", want: 25},
		"capped":      {query: "?limit=101", want: 100},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest("GET", "/"+tt.query, nil)
			assert.Equal(t, tt.want, listLimit(ctx))
		})
	}
}

func TestParseEventListFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("GET", "/?query=event-1&event_type=INSERT&outcome=processing&received_after=2026-07-22T01%3A02%3A03Z&received_before=2026-07-23T01%3A02%3A03Z", nil)

	filters, err := parseEventListFilters(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "event-1", filters.Query)
	assert.Equal(t, "INSERT", filters.EventType)
	assert.Equal(t, "processing", filters.Outcome)
	assert.Equal(t, time.Date(2026, 7, 22, 1, 2, 3, 0, time.UTC), *filters.ReceivedAfter)
	assert.Equal(t, time.Date(2026, 7, 23, 1, 2, 3, 0, time.UTC), *filters.ReceivedBefore)
}

func TestParseEventListFiltersRejectsInvalidValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for name, query := range map[string]string{
		"outcome":        "?outcome=unknown",
		"after":          "?received_after=yesterday",
		"before":         "?received_before=tomorrow",
		"reversed range": "?received_after=2026-07-23T00%3A00%3A00Z&received_before=2026-07-22T00%3A00%3A00Z",
	} {
		t.Run(name, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest("GET", "/"+query, nil)
			_, err := parseEventListFilters(ctx)
			assert.Error(t, err)
		})
	}
}

func TestEventRoutingStatuses(t *testing.T) {
	assert.Equal(t, []app.EventRoutingStatus{app.EventRoutingStatusMatched}, eventRoutingStatuses("ok"))
	assert.Equal(t, []app.EventRoutingStatus{app.EventRoutingStatusIgnored}, eventRoutingStatuses("ignored"))
	assert.Equal(t, []app.EventRoutingStatus{app.EventRoutingStatusRejected}, eventRoutingStatuses("rejected"))
	assert.Equal(t, []app.EventRoutingStatus{app.EventRoutingStatusRoutingFailed}, eventRoutingStatuses("failed"))
	assert.Equal(t, []app.EventRoutingStatus{app.EventRoutingStatusAccepted, app.EventRoutingStatusRouting}, eventRoutingStatuses("processing"))
	assert.Nil(t, eventRoutingStatuses(""))
}

func TestEventListCursor(t *testing.T) {
	receivedAt := time.Date(2026, 7, 22, 1, 2, 3, 4, time.UTC)
	encoded := encodeEventListCursor(app.TriggerEvent{ID: "event-1", ReceivedAt: receivedAt})
	cursor, err := decodeEventListCursor(encoded)
	assert.NoError(t, err)
	assert.Equal(t, "event-1", cursor.ID)
	assert.Equal(t, receivedAt, cursor.ReceivedAt)

	for _, malformed := range []string{"not-base64!", base64.RawURLEncoding.EncodeToString([]byte(`{}`)), base64.RawURLEncoding.EncodeToString([]byte(`{"received_at":"2026-07-22T01:02:03Z","id":"event-1","extra":true}`))} {
		_, err := decodeEventListCursor(malformed)
		assert.Error(t, err)
	}
}

func TestParseEventListOrder(t *testing.T) {
	for input, expected := range map[string]string{"": "desc", "desc": "desc", "DESC": "desc", " asc ": "asc"} {
		actual, err := parseEventListOrder(input)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	}
	_, err := parseEventListOrder("newest")
	assert.Error(t, err)
}

func TestEventCursorExpressionMatchesOrder(t *testing.T) {
	cursor := &eventListCursor{ReceivedAt: time.Date(2026, 7, 22, 1, 2, 3, 0, time.UTC), ID: "event-1"}
	asc := eventCursorExpression(cursor, "asc").(clause.Expr)
	desc := eventCursorExpression(cursor, "desc").(clause.Expr)
	assert.Contains(t, asc.SQL, "received_at > ?")
	assert.Contains(t, asc.SQL, "id > ?")
	assert.Contains(t, desc.SQL, "received_at < ?")
	assert.Contains(t, desc.SQL, "id < ?")
	assert.Equal(t, []any{cursor.ReceivedAt, cursor.ReceivedAt, cursor.ID}, asc.Vars)
}

func TestTriggerScopedEventSearchOnlyUsesEventAndExternalIDs(t *testing.T) {
	or, ok := triggerScopedEventSearch("GitHub-123").(clause.OrConditions)
	assert.True(t, ok)
	assert.Len(t, or.Exprs, 2)

	columns := make([]string, 0, 2)
	for _, expression := range or.Exprs {
		expr := expression.(clause.Expr)
		columns = append(columns, expr.Vars[0].(clause.Column).Name)
		assert.Equal(t, "%github-123%", expr.Vars[1])
	}
	assert.ElementsMatch(t, []string{"id", "external_id"}, columns)
}

func TestDispatchListCursor(t *testing.T) {
	createdAt := time.Date(2026, 7, 22, 1, 2, 3, 4, time.UTC)
	encoded := encodeDispatchListCursor(app.EventDispatch{ID: "dispatch-1", CreatedAt: createdAt})
	cursor, err := decodeDispatchListCursor(encoded)
	assert.NoError(t, err)
	assert.Equal(t, "dispatch-1", cursor.ID)
	assert.Equal(t, createdAt, cursor.CreatedAt)

	for _, malformed := range []string{"not-base64!", base64.RawURLEncoding.EncodeToString([]byte(`{}`)), base64.RawURLEncoding.EncodeToString([]byte(`{"created_at":"2026-07-22T01:02:03Z","id":"dispatch-1","extra":true}`))} {
		_, err := decodeDispatchListCursor(malformed)
		assert.Error(t, err)
	}
}

func TestEventResponseExposesRawBodyWithoutChangingEventJSON(t *testing.T) {
	event := app.TriggerEvent{RawBody: []byte(`{"secret":"redacted upstream"}`)}
	encoded, err := json.Marshal(eventResponse{TriggerEvent: event})
	assert.NoError(t, err)
	var response map[string]any
	assert.NoError(t, json.Unmarshal(encoded, &response))
	assert.NotContains(t, response, "raw_body_base64")
	assert.NotContains(t, response, "raw_body")

	encoded, err = json.Marshal(eventRawResponse{RawBodyBase64: "eyJzZWNyZXQiOiJyZWRhY3RlZCB1cHN0cmVhbSJ9"})
	assert.NoError(t, err)
	assert.JSONEq(t, `{"raw_body_base64":"eyJzZWNyZXQiOiJyZWRhY3RlZCB1cHN0cmVhbSJ9","raw_body_sha256":"","raw_body_size":0}`, string(encoded))
}

func TestEventSummaryIsCompactAndIncludesDispatchState(t *testing.T) {
	encoded, err := json.Marshal(eventSummaryResponse{
		ID:               "event-1",
		MatchCount:       1,
		WaiterMatchCount: 2,
		Dispatches: []dispatchSummary{{
			ID: "dispatch-1", Status: app.EventDispatchStatusDeadLettered, Error: "attempts exhausted",
		}},
	})
	assert.NoError(t, err)
	var response map[string]any
	assert.NoError(t, json.Unmarshal(encoded, &response))
	assert.NotContains(t, response, "payload")
	assert.NotContains(t, response, "raw_body")
	assert.NotContains(t, response, "match_explanations")
	assert.Equal(t, float64(1), response["match_count"])
	assert.Equal(t, float64(2), response["waiter_match_count"])
	assert.Equal(t, []any{map[string]any{"id": "dispatch-1", "status": "dead_lettered", "error": "attempts exhausted"}}, response["dispatches"])
}

func TestEventReplayable(t *testing.T) {
	assert.True(t, eventReplayable(app.EventRoutingStatusMatched))
	assert.True(t, eventReplayable(app.EventRoutingStatusIgnored))
	assert.True(t, eventReplayable(app.EventRoutingStatusRoutingFailed))
	assert.False(t, eventReplayable(app.EventRoutingStatusAccepted))
	assert.False(t, eventReplayable(app.EventRoutingStatusRouting))
	assert.False(t, eventReplayable(app.EventRoutingStatusRejected))
}
