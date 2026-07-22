package config

import "github.com/invopop/jsonschema"

type EventFilterConfig struct {
	Op    string `mapstructure:"op" toml:"op" jsonschema:"required,enum=eq"`
	Path  string `mapstructure:"path" toml:"path" jsonschema:"required,pattern=^/"`
	Value any    `mapstructure:"value" toml:"value" jsonschema:"required"`
}

func (c EventFilterConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "op", "filter operation; only exact equality is supported")
	addDescription(schema, "path", "nonempty RFC 6901 JSON Pointer into the event payload")
	addDescription(schema, "value", "JSON primitive value to compare")
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
	addDescription(schema, "filters", "optional exact-match payload filters")
	addDescription(schema, "match_all", "explicitly match every event from the source")
	addDescription(schema, "target", "app branch run target")
}

type EventsConfig struct {
	Rules []*EventRuleConfig `mapstructure:"rules,omitempty" toml:"rules,omitempty"`
}

func (c EventsConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "rules", "event automation rules")
}
