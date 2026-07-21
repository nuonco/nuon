package config

import "github.com/invopop/jsonschema"

type EventAutomationFilterConfig struct {
	Op    string `mapstructure:"op" toml:"op" jsonschema:"required,enum=eq"`
	Path  string `mapstructure:"path" toml:"path" jsonschema:"required,pattern=^/"`
	Value any    `mapstructure:"value" toml:"value" jsonschema:"required"`
}

func (c EventAutomationFilterConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "op", "filter operation; only exact equality is supported")
	addDescription(schema, "path", "nonempty RFC 6901 JSON Pointer into the event payload")
	addDescription(schema, "value", "JSON primitive value to compare")
}

type EventAutomationTargetConfig struct {
	Type      string `mapstructure:"type" toml:"type" jsonschema:"required,enum=app_branch_run"`
	AppBranch string `mapstructure:"app_branch" toml:"app_branch" jsonschema:"required"`
}

func (c EventAutomationTargetConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "type", "automation target type; only app_branch_run is supported")
	addDescription(schema, "app_branch", "declared app branch to run")
}

type EventAutomationConfig struct {
	Name        string                        `mapstructure:"name" toml:"name" jsonschema:"required"`
	EventSource string                        `mapstructure:"event_source" toml:"event_source" jsonschema:"required"`
	EventTypes  []string                      `mapstructure:"event_types" toml:"event_types" jsonschema:"required,minItems=1"`
	Filters     []EventAutomationFilterConfig `mapstructure:"filters,omitempty" toml:"filters,omitempty"`
	Target      *EventAutomationTargetConfig  `mapstructure:"target" toml:"target" jsonschema:"required"`
}

func (c EventAutomationConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "name", "unique event automation rule name")
	addDescription(schema, "event_source", "event source name")
	addDescription(schema, "event_types", "nonempty list of exact event type strings")
	addDescription(schema, "filters", "optional exact-match payload filters")
	addDescription(schema, "target", "app branch run target")
}
