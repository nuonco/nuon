package config

import "github.com/invopop/jsonschema"

type InstallsConfig struct {
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty" toml:"connected_repo,omitempty"`
	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty" toml:"public_repo,omitempty"`
}

func (c InstallsConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "connected_repo", "connected GitHub repo containing install config TOML files")
	addDescription(schema, "public_repo", "public git repo containing install config TOML files")
}

func (c *InstallsConfig) Validate() error {
	if c.ConnectedRepo != nil && c.PublicRepo != nil {
		return ErrConfig{
			Description: "installs: connected_repo and public_repo are mutually exclusive",
		}
	}
	return nil
}

func (c *InstallsConfig) parse() error {
	return nil
}
