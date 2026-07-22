package validate

import (
	"fmt"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/eventfilter"
)

func ValidateEventAutomations(cfg *config.AppConfig) error {
	if cfg.Events == nil {
		return nil
	}
	ruleNames := make(map[string]struct{}, len(cfg.Events.Rules))
	branchNames := make(map[string]struct{}, len(cfg.Branches)+1)
	if cfg.Branch != nil {
		branchNames[cfg.Branch.Name] = struct{}{}
	}
	for _, branch := range cfg.Branches {
		if branch != nil {
			branchNames[branch.Name] = struct{}{}
		}
	}

	for i, rule := range cfg.Events.Rules {
		prefix := fmt.Sprintf("events.rules[%d]", i)
		if rule == nil {
			return eventAutomationConfigErr("%s must not be null", prefix)
		}
		if rule.Name == "" {
			return eventAutomationConfigErr("%s.name is required", prefix)
		}
		if _, exists := ruleNames[rule.Name]; exists {
			return eventAutomationConfigErr("event automation rule name %q must be unique", rule.Name)
		}
		ruleNames[rule.Name] = struct{}{}
		if rule.Source == "" {
			return eventAutomationConfigErr("%s.source is required", prefix)
		}
		eventTypes := make(map[string]struct{}, len(rule.EventTypes))
		for _, eventType := range rule.EventTypes {
			if eventType == "" {
				return eventAutomationConfigErr("%s.event_types must not contain empty strings", prefix)
			}
			if _, exists := eventTypes[eventType]; exists {
				return eventAutomationConfigErr("%s.event_types contains duplicate %q", prefix, eventType)
			}
			eventTypes[eventType] = struct{}{}
		}
		hasPositiveFilter := false
		for j, filter := range rule.Filters {
			filterPrefix := fmt.Sprintf("%s.filters[%d]", prefix, j)
			compiledFilter := eventfilter.Filter{From: eventfilter.Source(filter.From), Path: filter.Path, Op: eventfilter.Operator(filter.Op), Value: filter.Value}
			if _, err := eventfilter.Compile(compiledFilter); err != nil {
				return eventAutomationConfigErr("%s is invalid: %v", filterPrefix, err)
			}
			hasPositiveFilter = hasPositiveFilter || eventfilter.IsPositive(compiledFilter.Op)
		}
		if len(rule.EventTypes) == 0 && !hasPositiveFilter && !rule.MatchAll {
			return eventAutomationConfigErr("%s must declare event_types, a positive filter, or match_all = true", prefix)
		}
		if rule.MatchAll && len(rule.EventTypes) != 0 {
			return eventAutomationConfigErr("%s.match_all cannot be combined with event_types", prefix)
		}
		if rule.Target == nil {
			return eventAutomationConfigErr("%s.target is required", prefix)
		}
		if rule.Target.Type != "app_branch_run" {
			return eventAutomationConfigErr("%s.target.type must be %q", prefix, "app_branch_run")
		}
		if rule.Target.AppBranch == "" {
			return eventAutomationConfigErr("%s.target.app_branch is required", prefix)
		}
		if _, exists := branchNames[rule.Target.AppBranch]; !exists {
			return eventAutomationConfigErr("%s.target.app_branch references undeclared app branch %q", prefix, rule.Target.AppBranch)
		}
	}
	return nil
}

func eventAutomationConfigErr(format string, args ...any) error {
	return config.ErrConfig{Description: fmt.Sprintf(format, args...)}
}
