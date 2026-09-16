package config

import (
	"fmt"

	"github.com/invopop/jsonschema"

	"github.com/nuonco/nuon/pkg/eventfilter"
)

type TriggerFilterConfig struct {
	From  string `json:"From" mapstructure:"from,omitempty" toml:"from,omitempty" jsonschema:"enum=payload,enum=headers"`
	Op    string `json:"Op" mapstructure:"op" toml:"op" jsonschema:"required,enum=eq,enum=neq,enum=in,enum=prefix,enum=suffix,enum=contains,enum=gt,enum=gte,enum=lt,enum=lte,enum=regex,enum=exists,enum=not_exists"`
	Path  string `json:"Path" mapstructure:"path" toml:"path" jsonschema:"required"`
	Value any    `json:"Value" mapstructure:"value,omitempty" toml:"value,omitempty"`
}

func (c TriggerFilterConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "from", "filter trigger: payload (default) or request headers")
	addDescription(schema, "op", "comparison operation")
	addDescription(schema, "path", "restricted JSONPath for payload filters, or a header name for header filters")
	addDescription(schema, "value", "comparison value; omitted for exists and not_exists")
}

type TriggerTargetConfig struct {
	Type      string            `json:"Type" mapstructure:"type" toml:"type" jsonschema:"required,enum=app_branch_run,enum=runbook"`
	AppBranch string            `json:"AppBranch" mapstructure:"app_branch,omitempty" toml:"app_branch,omitempty"`
	Runbook   string            `json:"Runbook" mapstructure:"runbook,omitempty" toml:"runbook,omitempty"`
	Install   string            `json:"Install" mapstructure:"install,omitempty" toml:"install,omitempty"`
	Inputs    map[string]string `json:"Inputs" mapstructure:"inputs,omitempty" toml:"inputs,omitempty"`
}

func (c TriggerTargetConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "type", "trigger target type")
	addDescription(schema, "app_branch", "declared app branch to run")
	addDescription(schema, "runbook", "declared runbook to run")
	addDescription(schema, "install", "app install name on which to run the runbook")
	addDescription(schema, "inputs", "runbook input names mapped to singular payload JSONPaths")
}

type TriggerRuleConfig struct {
	Name       string                `json:"Name" mapstructure:"name" toml:"name" jsonschema:"required"`
	Trigger    string                `json:"Trigger" mapstructure:"trigger" toml:"trigger" jsonschema:"required"`
	EventTypes []string              `json:"EventTypes" mapstructure:"event_types,omitempty" toml:"event_types,omitempty"`
	Filters    []TriggerFilterConfig `json:"Filters" mapstructure:"filters,omitempty" toml:"filters,omitempty"`
	MatchAll   bool                  `json:"MatchAll" mapstructure:"match_all,omitempty" toml:"match_all,omitempty"`
	Target     *TriggerTargetConfig  `json:"Target" mapstructure:"target" toml:"target" jsonschema:"required"`
}

func ValidateTriggerMatch(eventTypes []string, filters []TriggerFilterConfig, matchAll bool) error {
	if len(filters) > 20 {
		return fmt.Errorf("filters must contain at most 20 filters")
	}
	types := make(map[string]struct{}, len(eventTypes))
	for _, eventType := range eventTypes {
		if eventType == "" {
			return fmt.Errorf("event_types must not contain empty strings")
		}
		if _, exists := types[eventType]; exists {
			return fmt.Errorf("event_types contains duplicate %q", eventType)
		}
		types[eventType] = struct{}{}
	}
	hasPositiveFilter := false
	for i, filter := range filters {
		encodedValue, err := ToTOML(filter.Value)
		if err != nil || len(encodedValue) > 4096 {
			return fmt.Errorf("filters[%d].value must encode to at most 4096 bytes", i)
		}
		compiledFilter := eventfilter.Filter{From: eventfilter.Source(filter.From), Path: filter.Path, Op: eventfilter.Operator(filter.Op), Value: filter.Value}
		if _, err := eventfilter.Compile(compiledFilter); err != nil {
			return fmt.Errorf("filters[%d] is invalid: %v", i, err)
		}
		hasPositiveFilter = hasPositiveFilter || eventfilter.IsPositive(compiledFilter.Op)
	}
	if len(eventTypes) == 0 && !hasPositiveFilter && !matchAll {
		return fmt.Errorf("must declare event_types, a positive filter, or match_all = true")
	}
	if matchAll && len(eventTypes) != 0 {
		return fmt.Errorf("match_all cannot be combined with event_types")
	}
	return nil
}

func (c TriggerRuleConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "name", "unique trigger rule name")
	addDescription(schema, "trigger", "trigger name")
	addDescription(schema, "event_types", "optional list of exact event type strings")
	addDescription(schema, "filters", "optional ANDed payload or request-header predicates")
	addDescription(schema, "match_all", "explicit matching baseline, including for exclusion-only filters")
	addDescription(schema, "target", "app branch run or runbook target")
}

type TriggersConfig struct {
	Rules []*TriggerRuleConfig `json:"Rules" mapstructure:"rules,omitempty" toml:"rules,omitempty"`
}

func (c TriggersConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "rules", "trigger rules")
}
