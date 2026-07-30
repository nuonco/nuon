package activities

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestStaleRoutingGenerationCannotOverwriteEventProjection(t *testing.T) {
	db, err := gorm.Open(postgres.New(postgres.Config{DSN: "host=unused", PreferSimpleProtocol: true}), &gorm.Config{DisableAutomaticPing: true, SkipDefaultTransaction: true})
	require.NoError(t, err)
	db = db.Session(&gorm.Session{DryRun: true})

	event := app.TriggerEvent{ID: "ese00000000000000000000000"}
	r1 := "generation-r1"
	r2 := "generation-r2"
	event.RoutingGenerationToken = &r2

	result := eventGeneration(db, &event, r1).Updates(map[string]any{
		"routing_status": app.EventRoutingStatusRoutingFailed,
		"routing_error":  "R1 retry failed after R2 completed",
	})
	require.NoError(t, result.Error)
	require.Contains(t, strings.ToLower(result.Statement.SQL.String()), "routing_generation_token")
	require.Contains(t, result.Statement.Vars, &r1)
	require.NotContains(t, result.Statement.Vars, &r2)
}

func TestRuleMatchesTriggerEvent(t *testing.T) {
	payload, err := decodeTriggerEventPayload(json.RawMessage(`{"image":{"repository":"acme/api","tag":"v1"},"attempt":2}`))
	require.NoError(t, err)
	rule := &app.TriggerRule{
		EventTypes: pq.StringArray{"com.acme.image.pushed.v1"},
		Filters: []app.TriggerFilter{
			{Op: app.TriggerFilterTypeEq, Path: "$.image.repository", Value: "acme/api"},
			{Op: app.TriggerFilterTypeEq, Path: "$.attempt", Value: 2},
		},
	}

	require.True(t, ruleMatchesEvent(rule, "com.acme.image.pushed.v1", payload))
	require.False(t, ruleMatchesEvent(rule, "com.acme.image.deleted.v1", payload))
	rule.Filters[0].Value = "acme/other"
	require.False(t, ruleMatchesEvent(rule, "com.acme.image.pushed.v1", payload))
}

func TestWaiterMatchesOnlyEventsReceivedAfterActivation(t *testing.T) {
	activatedAt := time.Date(2026, time.July, 24, 12, 0, 0, 0, time.UTC)
	waiter := &app.EventRunbookWaiter{
		ActivatedAt: activatedAt,
		EventTypes:  pq.StringArray{"tag.updated"},
		Filters: []app.TriggerFilter{
			{Op: app.TriggerFilterTypeEq, Path: "$.repository", Value: "acme/api"},
		},
	}
	payload, err := decodeTriggerEventPayload(json.RawMessage(`{"repository":"acme/api"}`))
	require.NoError(t, err)

	event := &app.TriggerEvent{EventType: "tag.updated", ReceivedAt: activatedAt.Add(-time.Nanosecond)}
	require.False(t, waiterMatchesEvent(waiter, event, payload, nil))

	event.ReceivedAt = activatedAt
	require.True(t, waiterMatchesEvent(waiter, event, payload, nil))

	event.ReceivedAt = activatedAt.Add(time.Nanosecond)
	require.True(t, waiterMatchesEvent(waiter, event, payload, nil))
}

func TestRoutedEventStatusIncludesRunbookWaiterMatches(t *testing.T) {
	require.Equal(t, app.EventRoutingStatusIgnored, routedEventStatus(0, 0))
	require.Equal(t, app.EventRoutingStatusMatched, routedEventStatus(1, 0))
	require.Equal(t, app.EventRoutingStatusMatched, routedEventStatus(0, 1))
	require.Equal(t, app.EventRoutingStatusMatched, routedEventStatus(1, 1))
}

func TestRuleMatchesTriggerEventWithoutEventTypes(t *testing.T) {
	payload, err := decodeTriggerEventPayload(json.RawMessage(`{"ref":"main"}`))
	require.NoError(t, err)
	rule := &app.TriggerRule{Filters: []app.TriggerFilter{{Op: app.TriggerFilterTypeEq, Path: "$.ref", Value: "main"}}}
	require.True(t, ruleMatchesEvent(rule, "push", payload))
	require.True(t, ruleMatchesEvent(&app.TriggerRule{}, "push", payload))
}

func TestMapTriggerInputs(t *testing.T) {
	payload, err := decodeTriggerEventPayload(json.RawMessage(`{"tag":"v1","attempt":2,"forced":true,"empty":null,"image":{"tag":"v1"},"tags":["v1"]}`))
	require.NoError(t, err)

	mapped, err := mapTriggerInputs(map[string]string{
		"tag": "$.tag", "attempt": "$.attempt", "forced": "$.forced",
	}, payload)
	require.NoError(t, err)
	require.Equal(t, map[string]string{"tag": "v1", "attempt": "2", "forced": "true"}, mapped)

	for name, path := range map[string]string{
		"missing": "$.missing",
		"null":    "$.empty",
		"object":  "$.image",
		"array":   "$.tags",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := mapTriggerInputs(map[string]string{"input": path}, payload)
			require.Error(t, err)
		})
	}
}

func TestRuleMatchesTriggerEventWithWildcardAndHeaderFilters(t *testing.T) {
	payload, err := decodeTriggerEventPayload(json.RawMessage(`{"commits":[{"author":{"name":"Ada"}},{"author":{"name":"Grace"}}]}`))
	require.NoError(t, err)
	rule := &app.TriggerRule{Filters: []app.TriggerFilter{
		{Op: app.TriggerFilterTypeEq, Path: "$.commits[*].author.name", Value: "Grace"},
		{From: "headers", Op: app.TriggerFilterTypeEq, Path: "X-GitHub-Event", Value: "push"},
	}}
	require.True(t, ruleMatchesEvent(rule, "", payload, http.Header{"X-Github-Event": {"push"}}))
	require.False(t, ruleMatchesEvent(rule, "", payload, http.Header{"X-Github-Event": {"pull_request"}}))
	matched, explanation := evaluateRule(rule, "", payload, http.Header{"X-Github-Event": {"pull_request"}})
	require.False(t, matched)
	require.Len(t, explanation.Filters, 2)
	require.True(t, explanation.Filters[0].Matched)
	require.Equal(t, []any{"Ada", "Grace"}, explanation.Filters[0].Selected)
	require.False(t, explanation.Filters[1].Matched)
}

func TestActiveTriggerConfigIDsSelectsLatestNonPreviewPerApp(t *testing.T) {
	configs := []app.AppConfig{
		{ID: "preview", AppID: "app-a", Labeled: labels.Labeled{Labels: labels.Labels{"source": string(app.AppBranchRunTypeGitPreview)}}},
		{ID: "app-a-latest", AppID: "app-a"},
		{ID: "app-b-latest", AppID: "app-b"},
		{ID: "app-a-old", AppID: "app-a"},
	}
	require.Equal(t, map[string]string{"app-a": "app-a-latest", "app-b": "app-b-latest"}, app.ActiveTriggerConfigIDs(configs))
}
