package triggers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestBuildRule(t *testing.T) {
	validFrom := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	cfg := &config.TriggerRuleConfig{
		Name: "deploy", Trigger: "github", EventTypes: []string{"push"},
		Filters: []config.TriggerFilterConfig{{Op: "eq", Path: "$.ref", Value: "main"}},
		Target:  &config.TriggerTargetConfig{Type: "app_branch_run", AppBranch: "main"},
	}
	branchID := "branch"
	rule, err := buildRule(cfg, "org", "app", "config", "trigger", &branchID, nil, validFrom)
	require.NoError(t, err)
	require.Equal(t, "org", rule.OrgID)
	require.Equal(t, "config", rule.AppConfigID)
	require.Equal(t, app.TriggerTargetTypeAppBranchRun, rule.TargetType)
	require.True(t, rule.Enabled)
	require.True(t, rule.Force)
	require.False(t, rule.PlanOnly)
	require.Len(t, rule.ConfigHash, 64)
	require.Equal(t, []app.TriggerFilter{{Op: app.TriggerFilterTypeEq, Path: "$.ref", Value: "main"}}, rule.Filters)
}

func TestConfigHashIsNormalized(t *testing.T) {
	first := &config.TriggerRuleConfig{Name: "deploy", Trigger: "github", EventTypes: []string{"push", "pull_request"}, Filters: []config.TriggerFilterConfig{{Op: "eq", Path: "$.b", Value: true}, {Op: "eq", Path: "$.a", Value: 1}}, Target: &config.TriggerTargetConfig{Type: "app_branch_run", AppBranch: "main"}}
	second := &config.TriggerRuleConfig{Name: "deploy", Trigger: "github", EventTypes: []string{"pull_request", "push"}, Filters: []config.TriggerFilterConfig{{Op: "eq", Path: "$.a", Value: 1}, {Op: "eq", Path: "$.b", Value: true}}, Target: &config.TriggerTargetConfig{Type: "app_branch_run", AppBranch: "main"}}
	firstHash, err := configHash(first)
	require.NoError(t, err)
	secondHash, err := configHash(second)
	require.NoError(t, err)
	require.Equal(t, firstHash, secondHash)
}

func TestSyncValidatesBeforeDatabaseAccess(t *testing.T) {
	validTarget := &config.TriggerTargetConfig{Type: "app_branch_run", AppBranch: "main"}
	tests := map[string]*config.TriggerRuleConfig{
		"missing target": {
			Name: "deploy", Trigger: "github", MatchAll: true,
		},
		"invalid filter": {
			Name: "deploy", Trigger: "github", EventTypes: []string{"push"},
			Filters: []config.TriggerFilterConfig{{Op: "unknown", Path: "$..bad", Value: "main"}},
			Target:  validTarget,
		},
	}
	for name, rule := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := &config.AppConfig{
				Branch:   &config.AppBranchConfig{Name: "main"},
				Triggers: &config.TriggersConfig{Rules: []*config.TriggerRuleConfig{rule}},
			}
			require.Error(t, Sync(context.Background(), nil, cfg, "org", "app", "config"))
		})
	}
}

func TestReferencedTriggerNamesAreDeduplicatedAndSorted(t *testing.T) {
	rules := []*config.TriggerRuleConfig{
		{Trigger: "zebra"},
		{Trigger: "alpha"},
		{Trigger: "zebra"},
		{Trigger: "middle"},
	}
	require.Equal(t, []string{"alpha", "middle", "zebra"}, referencedTriggerNames(rules))
}
