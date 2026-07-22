package eventautomations

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestBuildRule(t *testing.T) {
	validFrom := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	cfg := &config.EventRuleConfig{
		Name: "deploy", Source: "github", EventTypes: []string{"push"},
		Filters: []config.EventFilterConfig{{Op: "eq", Path: "$.ref", Value: "main"}},
		Target:  &config.EventTargetConfig{Type: "app_branch_run", AppBranch: "main"},
	}
	rule, err := buildRule(cfg, "org", "app", "config", "source", "branch", validFrom)
	require.NoError(t, err)
	require.Equal(t, "org", rule.OrgID)
	require.Equal(t, "config", rule.AppConfigID)
	require.Equal(t, app.EventAutomationTargetTypeAppBranchRun, rule.TargetType)
	require.True(t, rule.Enabled)
	require.True(t, rule.Force)
	require.False(t, rule.PlanOnly)
	require.Len(t, rule.ConfigHash, 64)
	require.Equal(t, []app.EventAutomationFilter{{Op: app.EventAutomationFilterTypeEq, Path: "$.ref", Value: "main"}}, rule.Filters)
}

func TestConfigHashIsNormalized(t *testing.T) {
	first := &config.EventRuleConfig{Name: "deploy", Source: "github", EventTypes: []string{"push", "pull_request"}, Filters: []config.EventFilterConfig{{Op: "eq", Path: "$.b", Value: true}, {Op: "eq", Path: "$.a", Value: 1}}, Target: &config.EventTargetConfig{Type: "app_branch_run", AppBranch: "main"}}
	second := &config.EventRuleConfig{Name: "deploy", Source: "github", EventTypes: []string{"pull_request", "push"}, Filters: []config.EventFilterConfig{{Op: "eq", Path: "$.a", Value: 1}, {Op: "eq", Path: "$.b", Value: true}}, Target: &config.EventTargetConfig{Type: "app_branch_run", AppBranch: "main"}}
	firstHash, err := configHash(first)
	require.NoError(t, err)
	secondHash, err := configHash(second)
	require.NoError(t, err)
	require.Equal(t, firstHash, secondHash)
}
