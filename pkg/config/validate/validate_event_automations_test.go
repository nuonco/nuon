package validate

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
)

func validEventAutomationConfig() *config.AppConfig {
	return &config.AppConfig{
		Branch: &config.AppBranchConfig{Name: "main"},
		Events: &config.EventsConfig{Rules: []*config.EventRuleConfig{{
			Name: "deploy", Source: "github", EventTypes: []string{"push"},
			Filters: []config.EventFilterConfig{{Op: "eq", Path: `$['ref/name']`, Value: "main"}},
			Target:  &config.EventTargetConfig{Type: "app_branch_run", AppBranch: "main"},
		}}},
	}
}

func TestValidateEventAutomations(t *testing.T) {
	require.NoError(t, ValidateEventAutomations(validEventAutomationConfig()))

	tests := map[string]func(*config.AppConfig){
		"duplicate rule name": func(cfg *config.AppConfig) {
			cfg.Events.Rules = append(cfg.Events.Rules, cfg.Events.Rules[0])
		},
		"duplicate event type":  func(cfg *config.AppConfig) { cfg.Events.Rules[0].EventTypes = []string{"push", "push"} },
		"empty event type":      func(cfg *config.AppConfig) { cfg.Events.Rules[0].EventTypes = []string{""} },
		"invalid filter op":     func(cfg *config.AppConfig) { cfg.Events.Rules[0].Filters[0].Op = "unknown" },
		"empty path":            func(cfg *config.AppConfig) { cfg.Events.Rules[0].Filters[0].Path = "" },
		"unsupported recursive": func(cfg *config.AppConfig) { cfg.Events.Rules[0].Filters[0].Path = "$..bad" },
		"nonprimitive value":    func(cfg *config.AppConfig) { cfg.Events.Rules[0].Filters[0].Value = map[string]any{"x": true} },
		"invalid target type":   func(cfg *config.AppConfig) { cfg.Events.Rules[0].Target.Type = "action" },
		"unknown branch":        func(cfg *config.AppConfig) { cfg.Events.Rules[0].Target.AppBranch = "missing" },
		"implicit match all": func(cfg *config.AppConfig) {
			cfg.Events.Rules[0].EventTypes = nil
			cfg.Events.Rules[0].Filters = nil
		},
		"match all with discriminator": func(cfg *config.AppConfig) { cfg.Events.Rules[0].MatchAll = true },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := validEventAutomationConfig()
			mutate(cfg)
			require.Error(t, ValidateEventAutomations(cfg))
		})
	}
}

func TestValidateEventAutomationNegativeFilterGuard(t *testing.T) {
	negative := config.EventFilterConfig{Op: "neq", Path: "$.environment", Value: "development"}
	for name, configure := range map[string]func(*config.EventRuleConfig){
		"event types": func(rule *config.EventRuleConfig) { rule.EventTypes = []string{"push"} },
		"match all":   func(rule *config.EventRuleConfig) { rule.MatchAll = true },
	} {
		t.Run(name, func(t *testing.T) {
			cfg := validEventAutomationConfig()
			rule := cfg.Events.Rules[0]
			rule.EventTypes = nil
			rule.Filters = []config.EventFilterConfig{negative}
			configure(rule)
			require.NoError(t, ValidateEventAutomations(cfg))
		})
	}
	cfg := validEventAutomationConfig()
	cfg.Events.Rules[0].EventTypes = nil
	cfg.Events.Rules[0].Filters = []config.EventFilterConfig{negative}
	require.Error(t, ValidateEventAutomations(cfg))
}
