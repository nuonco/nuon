package activities

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestRuleMatchesAutomationEvent(t *testing.T) {
	payload, err := decodeAutomationPayload(json.RawMessage(`{"image":{"repository":"acme/api","tag":"v1"},"attempt":2}`))
	require.NoError(t, err)
	rule := &app.EventAutomationRule{
		EventTypes: pq.StringArray{"com.acme.image.pushed.v1"},
		Filters: []app.EventAutomationFilter{
			{Op: app.EventAutomationFilterTypeEq, Path: "$.image.repository", Value: "acme/api"},
			{Op: app.EventAutomationFilterTypeEq, Path: "$.attempt", Value: 2},
		},
	}

	require.True(t, ruleMatchesEvent(rule, "com.acme.image.pushed.v1", payload))
	require.False(t, ruleMatchesEvent(rule, "com.acme.image.deleted.v1", payload))
	rule.Filters[0].Value = "acme/other"
	require.False(t, ruleMatchesEvent(rule, "com.acme.image.pushed.v1", payload))
}

func TestRuleMatchesAutomationEventWithoutEventTypes(t *testing.T) {
	payload, err := decodeAutomationPayload(json.RawMessage(`{"ref":"main"}`))
	require.NoError(t, err)
	rule := &app.EventAutomationRule{Filters: []app.EventAutomationFilter{{Op: app.EventAutomationFilterTypeEq, Path: "$.ref", Value: "main"}}}
	require.True(t, ruleMatchesEvent(rule, "push", payload))
	require.True(t, ruleMatchesEvent(&app.EventAutomationRule{}, "push", payload))
}

func TestRuleMatchesAutomationEventWithWildcardAndHeaderFilters(t *testing.T) {
	payload, err := decodeAutomationPayload(json.RawMessage(`{"commits":[{"author":{"name":"Ada"}},{"author":{"name":"Grace"}}]}`))
	require.NoError(t, err)
	rule := &app.EventAutomationRule{Filters: []app.EventAutomationFilter{
		{Op: app.EventAutomationFilterTypeEq, Path: "$.commits[*].author.name", Value: "Grace"},
		{From: "headers", Op: app.EventAutomationFilterTypeEq, Path: "X-GitHub-Event", Value: "push"},
	}}
	require.True(t, ruleMatchesEvent(rule, "", payload, http.Header{"X-Github-Event": {"push"}}))
	require.False(t, ruleMatchesEvent(rule, "", payload, http.Header{"X-Github-Event": {"pull_request"}}))
}

func TestActiveAutomationConfigIDsSelectsLatestNonPreviewPerApp(t *testing.T) {
	configs := []app.AppConfig{
		{ID: "preview", AppID: "app-a", Labeled: labels.Labeled{Labels: labels.Labels{"source": string(app.AppBranchRunTypeGitPreview)}}},
		{ID: "app-a-latest", AppID: "app-a"},
		{ID: "app-b-latest", AppID: "app-b"},
		{ID: "app-a-old", AppID: "app-a"},
	}
	require.Equal(t, map[string]string{"app-a": "app-a-latest", "app-b": "app-b-latest"}, activeAutomationConfigIDs(configs))
}
