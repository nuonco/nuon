package validate

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
)

func validTriggerConfig() *config.AppConfig {
	return &config.AppConfig{
		Branch: &config.AppBranchConfig{Name: "main"},
		Triggers: &config.TriggersConfig{Rules: []*config.TriggerRuleConfig{{
			Name: "deploy", Trigger: "github", EventTypes: []string{"push"},
			Filters: []config.TriggerFilterConfig{{Op: "eq", Path: `$['ref/name']`, Value: "main"}},
			Target:  &config.TriggerTargetConfig{Type: "app_branch_run", AppBranch: "main"},
		}}},
	}
}

func TestValidateTriggers(t *testing.T) {
	require.NoError(t, ValidateTriggers(validTriggerConfig()))

	tests := map[string]func(*config.AppConfig){
		"duplicate rule name": func(cfg *config.AppConfig) {
			cfg.Triggers.Rules = append(cfg.Triggers.Rules, cfg.Triggers.Rules[0])
		},
		"duplicate event type":  func(cfg *config.AppConfig) { cfg.Triggers.Rules[0].EventTypes = []string{"push", "push"} },
		"empty event type":      func(cfg *config.AppConfig) { cfg.Triggers.Rules[0].EventTypes = []string{""} },
		"invalid filter op":     func(cfg *config.AppConfig) { cfg.Triggers.Rules[0].Filters[0].Op = "unknown" },
		"empty path":            func(cfg *config.AppConfig) { cfg.Triggers.Rules[0].Filters[0].Path = "" },
		"unsupported recursive": func(cfg *config.AppConfig) { cfg.Triggers.Rules[0].Filters[0].Path = "$..bad" },
		"nonprimitive value":    func(cfg *config.AppConfig) { cfg.Triggers.Rules[0].Filters[0].Value = map[string]any{"x": true} },
		"invalid target type":   func(cfg *config.AppConfig) { cfg.Triggers.Rules[0].Target.Type = "action" },
		"unknown branch":        func(cfg *config.AppConfig) { cfg.Triggers.Rules[0].Target.AppBranch = "missing" },
		"implicit match all": func(cfg *config.AppConfig) {
			cfg.Triggers.Rules[0].EventTypes = nil
			cfg.Triggers.Rules[0].Filters = nil
		},
		"match all with discriminator": func(cfg *config.AppConfig) { cfg.Triggers.Rules[0].MatchAll = true },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := validTriggerConfig()
			mutate(cfg)
			require.Error(t, ValidateTriggers(cfg))
		})
	}
}

func TestValidateRunbookTrigger(t *testing.T) {
	cfg := validTriggerConfig()
	cfg.Runbooks = []*config.RunbookConfig{{
		Name: "deploy-image",
		Inputs: []config.RunbookInput{
			{Name: "image_tag", Required: true},
			{Name: "environment", Required: true, Default: "production"},
		},
	}}
	cfg.Triggers.Rules[0].Target = &config.TriggerTargetConfig{
		Type: "runbook", Runbook: "deploy-image", Install: "production",
		Inputs: map[string]string{"image_tag": "$.tag"},
	}
	require.NoError(t, ValidateTriggers(cfg))

	tests := map[string]func(*config.TriggerTargetConfig){
		"unknown runbook":          func(target *config.TriggerTargetConfig) { target.Runbook = "missing" },
		"unknown input":            func(target *config.TriggerTargetConfig) { target.Inputs["missing"] = "$.tag" },
		"missing required mapping": func(target *config.TriggerTargetConfig) { delete(target.Inputs, "image_tag") },
		"wildcard mapping":         func(target *config.TriggerTargetConfig) { target.Inputs["image_tag"] = "$.tags[*]" },
		"recursive mapping":        func(target *config.TriggerTargetConfig) { target.Inputs["image_tag"] = "$..tag" },
		"branch field":             func(target *config.TriggerTargetConfig) { target.AppBranch = "main" },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			testCfg := validTriggerConfig()
			testCfg.Runbooks = cfg.Runbooks
			testCfg.Triggers.Rules[0].Target = &config.TriggerTargetConfig{
				Type: "runbook", Runbook: "deploy-image", Install: "production",
				Inputs: map[string]string{"image_tag": "$.tag"},
			}
			mutate(testCfg.Triggers.Rules[0].Target)
			require.Error(t, ValidateTriggers(testCfg))
		})
	}
}

func TestValidateTriggerNegativeFilterGuard(t *testing.T) {
	negative := config.TriggerFilterConfig{Op: "neq", Path: "$.environment", Value: "development"}
	for name, configure := range map[string]func(*config.TriggerRuleConfig){
		"event types": func(rule *config.TriggerRuleConfig) { rule.EventTypes = []string{"push"} },
		"match all":   func(rule *config.TriggerRuleConfig) { rule.MatchAll = true },
	} {
		t.Run(name, func(t *testing.T) {
			cfg := validTriggerConfig()
			rule := cfg.Triggers.Rules[0]
			rule.EventTypes = nil
			rule.Filters = []config.TriggerFilterConfig{negative}
			configure(rule)
			require.NoError(t, ValidateTriggers(cfg))
		})
	}
	cfg := validTriggerConfig()
	cfg.Triggers.Rules[0].EventTypes = nil
	cfg.Triggers.Rules[0].Filters = []config.TriggerFilterConfig{negative}
	require.Error(t, ValidateTriggers(cfg))
}
