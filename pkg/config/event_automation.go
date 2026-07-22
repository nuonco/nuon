package config

import "github.com/invopop/jsonschema"

type EventFilterConfig struct {
	From  string `mapstructure:"from,omitempty" toml:"from,omitempty" jsonschema:"enum=payload,enum=headers"`
	Op    string `mapstructure:"op" toml:"op" jsonschema:"required,enum=eq,enum=neq,enum=in,enum=prefix,enum=suffix,enum=contains,enum=gt,enum=gte,enum=lt,enum=lte,enum=regex,enum=exists,enum=not_exists"`
	Path  string `mapstructure:"path" toml:"path" jsonschema:"required"`
	Value any    `mapstructure:"value,omitempty" toml:"value,omitempty"`
}

func (c EventFilterConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "from", "filter source: payload (default) or request headers")
	addDescription(schema, "op", "comparison operation")
	addDescription(schema, "path", "restricted JSONPath for payload filters, or a header name for header filters")
	addDescription(schema, "value", "comparison value; omitted for exists and not_exists")
}

type EventTargetConfig struct {
	Type      string `mapstructure:"type" toml:"type" jsonschema:"required,enum=app_branch_run"`
	AppBranch string `mapstructure:"app_branch" toml:"app_branch" jsonschema:"required"`
}

func (c EventTargetConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "type", "automation target type; only app_branch_run is supported")
	addDescription(schema, "app_branch", "declared app branch to run")
}

type EventRuleConfig struct {
	Name       string              `mapstructure:"name" toml:"name" jsonschema:"required"`
	Source     string              `mapstructure:"source" toml:"source" jsonschema:"required"`
	EventTypes []string            `mapstructure:"event_types,omitempty" toml:"event_types,omitempty"`
	Filters    []EventFilterConfig `mapstructure:"filters,omitempty" toml:"filters,omitempty"`
	MatchAll   bool                `mapstructure:"match_all,omitempty" toml:"match_all,omitempty"`
	Target     *EventTargetConfig  `mapstructure:"target" toml:"target" jsonschema:"required"`
}

func (c EventRuleConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "name", "unique event automation rule name")
	addDescription(schema, "source", "event source name")
	addDescription(schema, "event_types", "optional list of exact event type strings")
	addDescription(schema, "filters", "optional ANDed payload or request-header predicates")
	addDescription(schema, "match_all", "explicit matching baseline, including for exclusion-only filters")
	addDescription(schema, "target", "app branch run target")
}

type EventsConfig struct {
	Rules []*EventRuleConfig `mapstructure:"rules,omitempty" toml:"rules,omitempty"`
}

func (c EventsConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "rules", "event automation rules")
}
