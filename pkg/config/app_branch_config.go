package config

import "github.com/invopop/jsonschema"

type AppBranchConfig struct {
	Name          string               `mapstructure:"name" jsonschema:"required"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty" jsonschema:"required"`
}

func (c AppBranchConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "name", "name of the app branch")
	addDescription(schema, "connected_repo", "git branch the app config tracks")
}

func (c *AppBranchConfig) Validate() error {
	return nil
}

func (c *AppBranchConfig) parse() error {
	return nil
}
