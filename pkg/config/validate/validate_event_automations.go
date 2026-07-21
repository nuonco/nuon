package validate

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nuonco/nuon/pkg/config"
)

func ValidateEventAutomations(cfg *config.AppConfig) error {
	ruleNames := make(map[string]struct{}, len(cfg.EventAutomations))
	branchNames := make(map[string]struct{}, len(cfg.Branches)+1)
	if cfg.Branch != nil {
		branchNames[cfg.Branch.Name] = struct{}{}
	}
	for _, branch := range cfg.Branches {
		if branch != nil {
			branchNames[branch.Name] = struct{}{}
		}
	}

	for i, rule := range cfg.EventAutomations {
		prefix := fmt.Sprintf("event_automations[%d]", i)
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
		if rule.EventSource == "" {
			return eventAutomationConfigErr("%s.event_source is required", prefix)
		}
		if len(rule.EventTypes) == 0 {
			return eventAutomationConfigErr("%s.event_types must contain at least one event type", prefix)
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
		for j, filter := range rule.Filters {
			filterPrefix := fmt.Sprintf("%s.filters[%d]", prefix, j)
			if filter.Op != "eq" {
				return eventAutomationConfigErr("%s.op must be %q", filterPrefix, "eq")
			}
			if !validJSONPointer(filter.Path) {
				return eventAutomationConfigErr("%s.path must be a nonempty RFC 6901 JSON Pointer", filterPrefix)
			}
			if !isJSONPrimitive(filter.Value) {
				return eventAutomationConfigErr("%s.value must be a JSON primitive", filterPrefix)
			}
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

func validJSONPointer(pointer string) bool {
	if !strings.HasPrefix(pointer, "/") {
		return false
	}
	for i := 0; i < len(pointer); i++ {
		if pointer[i] == '~' {
			if i+1 >= len(pointer) || (pointer[i+1] != '0' && pointer[i+1] != '1') {
				return false
			}
			i++
		}
	}
	return true
}

func isJSONPrimitive(value any) bool {
	switch value.(type) {
	case nil, bool, string, json.Number,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	default:
		return false
	}
}

func eventAutomationConfigErr(format string, args ...any) error {
	return config.ErrConfig{Description: fmt.Sprintf(format, args...)}
}
