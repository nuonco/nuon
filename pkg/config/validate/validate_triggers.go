package validate

import (
	"fmt"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/eventfilter"
)

func ValidateTriggers(cfg *config.AppConfig) error {
	if cfg.Triggers == nil {
		return nil
	}
	if len(cfg.Triggers.Rules) > 100 {
		return triggerConfigErr("triggers.rules must contain at most 100 rules")
	}
	ruleNames := make(map[string]struct{}, len(cfg.Triggers.Rules))
	branchNames := make(map[string]struct{}, len(cfg.Branches)+1)
	runbooks := make(map[string]*config.RunbookConfig, len(cfg.Runbooks))
	for _, runbook := range cfg.Runbooks {
		if runbook != nil {
			runbooks[runbook.Name] = runbook
		}
	}
	if cfg.Branch != nil {
		branchNames[cfg.Branch.Name] = struct{}{}
	}
	for _, branch := range cfg.Branches {
		if branch != nil {
			branchNames[branch.Name] = struct{}{}
		}
	}

	for i, rule := range cfg.Triggers.Rules {
		prefix := fmt.Sprintf("triggers.rules[%d]", i)
		if rule == nil {
			return triggerConfigErr("%s must not be null", prefix)
		}
		if rule.Name == "" {
			return triggerConfigErr("%s.name is required", prefix)
		}
		if _, exists := ruleNames[rule.Name]; exists {
			return triggerConfigErr("trigger rule name %q must be unique", rule.Name)
		}
		ruleNames[rule.Name] = struct{}{}
		if rule.Trigger == "" {
			return triggerConfigErr("%s.trigger is required", prefix)
		}
		if err := config.ValidateTriggerMatch(rule.EventTypes, rule.Filters, rule.MatchAll); err != nil {
			return triggerConfigErr("%s.%v", prefix, err)
		}
		if rule.Target == nil {
			return triggerConfigErr("%s.target is required", prefix)
		}
		switch rule.Target.Type {
		case "app_branch_run":
			if rule.Target.AppBranch == "" {
				return triggerConfigErr("%s.target.app_branch is required", prefix)
			}
			if rule.Target.Runbook != "" || rule.Target.Install != "" || len(rule.Target.Inputs) != 0 {
				return triggerConfigErr("%s.target app_branch_run cannot declare runbook, install, or inputs", prefix)
			}
			if _, exists := branchNames[rule.Target.AppBranch]; !exists {
				return triggerConfigErr("%s.target.app_branch references undeclared app branch %q", prefix, rule.Target.AppBranch)
			}
		case "runbook":
			if rule.Target.Runbook == "" || rule.Target.Install == "" {
				return triggerConfigErr("%s.target.runbook and target.install are required", prefix)
			}
			if rule.Target.AppBranch != "" {
				return triggerConfigErr("%s.target runbook cannot declare app_branch", prefix)
			}
			if len(rule.Target.Inputs) > 50 {
				return triggerConfigErr("%s.target.inputs must contain at most 50 mappings", prefix)
			}
			runbook, ok := runbooks[rule.Target.Runbook]
			if !ok {
				return triggerConfigErr("%s.target.runbook references undeclared runbook %q", prefix, rule.Target.Runbook)
			}
			declared := make(map[string]config.RunbookInput, len(runbook.Inputs))
			for _, input := range runbook.Inputs {
				declared[input.Name] = input
			}
			for name, path := range rule.Target.Inputs {
				if _, ok := declared[name]; !ok {
					return triggerConfigErr("%s.target.inputs references undeclared input %q", prefix, name)
				}
				if _, err := eventfilter.ParsePath(path, false); err != nil {
					return triggerConfigErr("%s.target.inputs[%q] is invalid: %v", prefix, name, err)
				}
			}
			for _, input := range runbook.Inputs {
				if input.Required && (input.Default == nil || fmt.Sprint(input.Default) == "") {
					if _, ok := rule.Target.Inputs[input.Name]; !ok {
						return triggerConfigErr("%s.target.inputs must map required input %q", prefix, input.Name)
					}
				}
			}
		default:
			return triggerConfigErr("%s.target.type must be app_branch_run or runbook", prefix)
		}
	}
	return nil
}

func triggerConfigErr(format string, args ...any) error {
	return config.ErrConfig{Description: fmt.Sprintf(format, args...)}
}
