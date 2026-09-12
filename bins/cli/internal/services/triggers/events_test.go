package triggers

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func TestEvaluateRules(t *testing.T) {
	event := &models.TriggerEvent{
		ID: "evt1", TriggerName: "github", EventType: "push",
		Payload: json.RawMessage(`{"ref":"refs/heads/main","count":2}`),
		Headers: map[string][]string{"X-Github-Event": {"push"}},
	}
	cfg := &config.AppConfig{Triggers: &config.TriggersConfig{Rules: []*config.TriggerRuleConfig{
		{Name: "deploy", Trigger: "github", EventTypes: []string{"push", "pull_request"}, Filters: []config.TriggerFilterConfig{
			{Path: "$.ref", Op: "eq", Value: "refs/heads/main"},
			{From: "headers", Path: "x-github-event", Op: "eq", Value: "push"},
			{Path: "$.count", Op: "gt", Value: 1},
		}},
		{Name: "other trigger", Trigger: "stripe", MatchAll: true},
	}}}
	result, err := EvaluateRules(event, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rules) != 1 || !result.Rules[0].Matched {
		t.Fatalf("unexpected rules: %#v", result.Rules)
	}
	rule := result.Rules[0]
	if !rule.EventTypes[0].Matched || rule.EventTypes[1].Matched {
		t.Fatalf("unexpected event type results: %#v", rule.EventTypes)
	}
	for _, filter := range rule.Filters {
		if !filter.Matched || len(filter.Selected) != 1 {
			t.Fatalf("unexpected filter result: %#v", filter)
		}
	}
}

func TestEvaluateRulesRequiresTriggerName(t *testing.T) {
	_, err := EvaluateRules(&models.TriggerEvent{Payload: json.RawMessage(`{}`)}, &config.AppConfig{})
	if err == nil || !strings.Contains(err.Error(), "trigger name is absent") {
		t.Fatalf("expected absent trigger name error, got %v", err)
	}
}

func TestEvaluateRulesSkipsTriggerMismatch(t *testing.T) {
	event := &models.TriggerEvent{TriggerName: "github", Payload: json.RawMessage(`{}`)}
	cfg := &config.AppConfig{Triggers: &config.TriggersConfig{Rules: []*config.TriggerRuleConfig{
		{Name: "stripe rule", Trigger: "stripe", MatchAll: true},
	}}}
	result, err := EvaluateRules(event, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Rules) != 0 {
		t.Fatalf("expected mismatched trigger rule to be skipped, got %#v", result.Rules)
	}
}

func TestFlattenPaths(t *testing.T) {
	event := &models.TriggerEvent{
		Payload: json.RawMessage(`{"simple":{"items":[true,2]},"bad key":{"quote'\\":null},"empty":{},"none":[]}`),
		Headers: map[string][]string{"X-Signature": {"one", "two"}},
	}
	rows, err := FlattenPaths(event)
	if err != nil {
		t.Fatal(err)
	}
	want := []PathRow{
		{From: "headers", Path: "X-Signature", Value: []string{"one", "two"}},
		{From: "payload", Path: "$.empty", Value: map[string]any{}},
		{From: "payload", Path: "$.none", Value: []any{}},
		{From: "payload", Path: "$.simple.items[0]", Value: true},
		{From: "payload", Path: "$.simple.items[1]", Value: json.Number("2")},
		{From: "payload", Path: `$["bad key"]["quote'\\"]`, Value: nil},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("rows mismatch\ngot:  %#v\nwant: %#v", rows, want)
	}
}

func TestFlattenPayloadBoundsPathsAndDepth(t *testing.T) {
	wide := make(map[string]any, eventPathMaxCount+1)
	for idx := 0; idx <= eventPathMaxCount; idx++ {
		wide[fmt.Sprintf("key-%04d", idx)] = idx
	}
	rows := flattenPayload(wide)
	require.Len(t, rows, eventPathMaxCount+1)
	require.Equal(t, eventPathTruncatedValue, rows[len(rows)-1].Value)

	var deep any = "leaf"
	for idx := 0; idx <= eventPathMaxDepth; idx++ {
		deep = map[string]any{"child": deep}
	}
	rows = flattenPayload(deep)
	require.Equal(t, []PathRow{{From: "payload", Path: "…", Value: eventPathTruncatedValue}}, rows)
}

func TestTriggerEventDispatchSummaryIsDeterministicAndIncludesFailures(t *testing.T) {
	dispatches := []*models.TriggerEventDispatch{
		{ID: "dispatch-b", Status: "triggered"},
		{ID: "dispatch-a", Status: "dead_lettered", Error: "attempts exhausted"},
	}
	summaries := make([]models.TriggerEventDispatchSummary, len(dispatches))
	for i, dispatch := range dispatches {
		summaries[i] = models.TriggerEventDispatchSummary{ID: dispatch.ID, Status: dispatch.Status, Error: dispatch.Error}
	}
	want := "dispatch-a=dead_lettered (attempts exhausted), dispatch-b=triggered"
	if got := triggerEventDispatchSummary(summaries); got != want {
		t.Fatalf("triggerEventDispatchSummary() = %q, want %q", got, want)
	}
}

func TestSanitizeHumanTextEscapesTerminalControls(t *testing.T) {
	input := "push\x1b[31mred\x1b[0m\x1b]8;;https://evil.example\x07link\x1b]8;;\x1b\\\n\u0085end"
	want := `push\x1b[31mred\x1b[0m\x1b]8;;https://evil.example\x07link\x1b]8;;\x1b\\x0a\x85end`
	require.Equal(t, want, sanitizeHumanText(input))
	require.NotContains(t, sanitizeHumanText(input), "\x1b")
}

func TestTailSeenBoundsTerminalEventsAndRetainsTransitions(t *testing.T) {
	seen := newTailSeen(2)
	record := func(id, routing, dispatch string) {
		event := &models.TriggerEventSummary{ID: id, RoutingStatus: routing}
		if dispatch != "" {
			event.Dispatches = []models.TriggerEventDispatchSummary{{ID: "dispatch-" + id, Status: dispatch}}
		}
		seen.record(event, routing+"|"+dispatch)
	}

	record("active", "matched", "pending")
	record("terminal-1", "matched", "triggered")
	record("terminal-2", "ignored", "")
	record("terminal-3", "routing_failed", "")
	require.Len(t, seen.fingerprints, 3)
	require.Contains(t, seen.fingerprints, "active")
	require.NotContains(t, seen.fingerprints, "terminal-1")

	record("active", "matched", "dead_lettered")
	require.True(t, seen.terminal["active"])
	record("new-active", "routing", "")
	record("terminal-4", "rejected", "")
	require.Contains(t, seen.fingerprints, "new-active")
	require.NotContains(t, seen.fingerprints, "terminal-2")
}

type tailAPI struct {
	cancel     context.CancelFunc
	listCalls  int
	rawCalls   int
	transition bool
}

func (a *tailAPI) ListTriggerEvents(_ context.Context, _ int, _ string) ([]*models.TriggerEventSummary, error) {
	a.listCalls++
	status := "pending"
	if a.transition && a.listCalls > 1 {
		status = "dead_lettered"
	}
	if a.listCalls == 2 {
		defer a.cancel()
	}
	return []*models.TriggerEventSummary{{ID: "event-1", RoutingStatus: "matched", ReceivedAt: time.Now(), Dispatches: []models.TriggerEventDispatchSummary{{ID: "dispatch-1", Status: status}}}}, nil
}

func (a *tailAPI) ListTriggerEventsPage(ctx context.Context, limit int, source, _ string) (*models.TriggerEventPage, error) {
	items, err := a.ListTriggerEvents(ctx, limit, source)
	return &models.TriggerEventPage{Items: items}, err
}
func (a *tailAPI) SearchTriggerEvents(ctx context.Context, filters models.TriggerEventListQuery) (*models.TriggerEventPage, error) {
	return a.ListTriggerEventsPage(ctx, filters.Limit, filters.Trigger, filters.Cursor)
}

func (a *tailAPI) GetTriggerEventRaw(context.Context, string) (*models.TriggerEventRaw, error) {
	a.rawCalls++
	return &models.TriggerEventRaw{RawBodyBase64: "e30="}, nil
}

func (a *tailAPI) GetTriggerEvent(context.Context, string) (*models.TriggerEvent, error) {
	return nil, nil
}

func (a *tailAPI) ReplayTriggerEvent(context.Context, string) (*models.TriggerEventReplayResponse, error) {
	return nil, nil
}
func (a *tailAPI) GetTriggerEventDispatch(context.Context, string) (*models.TriggerEventDispatch, error) {
	return nil, nil
}
func (a *tailAPI) RetryTriggerEventDispatch(context.Context, string) (*models.TriggerEventDispatchRetryResponse, error) {
	return nil, nil
}
func (a *tailAPI) ListTriggerEventDispatchesPage(context.Context, int, string, string) (*models.TriggerEventDispatchPage, error) {
	return nil, nil
}
func (a *tailAPI) CreateTrigger(context.Context, *models.TriggerCreateRequest) (*models.TriggerCredentialResponse, error) {
	return nil, nil
}
func (a *tailAPI) ListTriggers(context.Context) ([]*models.Trigger, error) {
	return []*models.Trigger{{ID: "trigger-1", Name: "github"}, {ID: "trigger-2", Name: "dup"}, {ID: "trigger-3", Name: "dup"}}, nil
}
func (a *tailAPI) GetTrigger(context.Context, string) (*models.Trigger, error) {
	return nil, nil
}
func (a *tailAPI) RotateTriggerSecret(context.Context, string) (*models.TriggerCredentialResponse, error) {
	return nil, nil
}
func (a *tailAPI) EnableTrigger(context.Context, string) (*models.Trigger, error) {
	return nil, nil
}
func (a *tailAPI) DisableTrigger(context.Context, string) (*models.Trigger, error) {
	return nil, nil
}
func (a *tailAPI) RevokeTriggerSecret(context.Context, string, string) (*models.TriggerRevokeResponse, error) {
	return nil, nil
}
func (a *tailAPI) RevealTriggerSecret(context.Context, string, string) (*models.TriggerSecretRevealResponse, error) {
	return nil, nil
}
func (a *tailAPI) GetTriggerIngressURL(context.Context, string) (*models.TriggerIngressURLResponse, error) {
	return nil, nil
}
func (a *tailAPI) RotateTriggerIngressURL(context.Context, string) (*models.TriggerCredentialResponse, error) {
	return nil, nil
}
func (a *tailAPI) DeleteTrigger(context.Context, string, bool) error { return nil }

func TestTailSuppressesHistoryAndEmitsDispatchTransitions(t *testing.T) {
	for _, test := range []struct {
		name       string
		transition bool
		rawCalls   int
	}{
		{name: "unchanged", rawCalls: 0},
		{name: "dispatch transition", transition: true, rawCalls: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			api := &tailAPI{cancel: cancel, transition: test.transition}
			require.NoError(t, (&Service{api: api}).Tail(ctx, "trigger-1", time.Millisecond, true))
			require.Equal(t, 2, api.listCalls)
			require.Equal(t, test.rawCalls, api.rawCalls)
		})
	}
}

func TestResolveTriggerID(t *testing.T) {
	service := &Service{api: &tailAPI{}}
	for _, test := range []struct {
		trigger string
		want    string
		wantErr string
	}{
		{trigger: "trigger-1", want: "trigger-1"},
		{trigger: "github", want: "trigger-1"},
		{trigger: "dup", wantErr: "multiple triggers are named"},
		{trigger: "missing", wantErr: "not found"},
		{trigger: "", wantErr: "--trigger is required"},
	} {
		got, err := service.resolveTriggerID(context.Background(), test.trigger)
		if test.wantErr != "" {
			require.ErrorContains(t, err, test.wantErr, test.trigger)
			continue
		}
		require.NoError(t, err, test.trigger)
		require.Equal(t, test.want, got, test.trigger)
	}
}

type pagedTailAPI struct {
	tailAPI
	pages map[string]*models.TriggerEventPage
	calls []string
}

func (a *pagedTailAPI) ListTriggerEventsPage(_ context.Context, _ int, _, cursor string) (*models.TriggerEventPage, error) {
	a.calls = append(a.calls, cursor)
	return a.pages[cursor], nil
}

func TestEventsSinceSeenFollowsPagesUntilKnownEvent(t *testing.T) {
	first := make([]*models.TriggerEventSummary, 100)
	for i := range first {
		first[i] = &models.TriggerEventSummary{ID: fmt.Sprintf("new-%03d", i)}
	}
	second := make([]*models.TriggerEventSummary, 51)
	for i := 0; i < 50; i++ {
		second[i] = &models.TriggerEventSummary{ID: fmt.Sprintf("new-%03d", i+100)}
	}
	second[50] = &models.TriggerEventSummary{ID: "known"}
	api := &pagedTailAPI{pages: map[string]*models.TriggerEventPage{
		"":       {Items: first, NextCursor: "page-2"},
		"page-2": {Items: second, NextCursor: "page-3"},
	}}
	seen := newTailSeen(tailRecentTerminalLimit)
	seen.fingerprints["known"] = "matched|—"
	events, err := (&Service{api: api}).eventsSinceSeen(context.Background(), "trigger", seen, true)
	require.NoError(t, err)
	require.Len(t, events, 151)
	require.Equal(t, []string{"", "page-2"}, api.calls)
}

func TestMemberPath(t *testing.T) {
	tests := map[string]string{"name": ".name", "_id2": "._id2", "two words": `["two words"]`, "9lives": `["9lives"]`, "line\nbreak": `["line\nbreak"]`}
	for input, want := range tests {
		if got := memberPath(input); got != want {
			t.Errorf("memberPath(%q) = %q, want %q", input, got, want)
		}
	}
}
