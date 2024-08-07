package config


// These structs are only used for jsonschema validation for source(s) validation
type AppSourceConfig struct {
	Source string `mapstructure:"source,omitempty" jsonschema:"required"`
}

type AppSourcesConfig struct {
	Source string `mapstructure:"source,omitempty" jsonschema:"oneof_required=source"`
	Sources []string `mapstructure:"sources,omitempty" jsonschema:"oneof_required=sources"`
}
