package config

import (
	"github.com/invopop/jsonschema"
)

type PublicRepoConfig struct {
	Repo      string `mapstructure:"repo,omitempty" jsonschema:"required"`
	Directory string `mapstructure:"directory,omitempty" jsonschema:"required"`
	Branch    string `mapstructure:"branch,omitempty" jsonschema:"required"`
}

func (p PublicRepoConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("repo").Short("repository URL").Required().
		Long("HTTPS URL to the public Git repository").
		Example("https://github.com/user/repo.git").
		Example("https://github.com/user/terraform-modules.git").
		Field("directory").Short("directory path").Required().
		Long("Path within the repository to the configuration files").
		Example("terraform").
		Example("infra/terraform").
		Field("branch").Short("Git branch").Required().
		Long("Git branch to checkout and use for deployments").
		Example("main").
		Example("develop").
		Example("production")
}

type ConnectedRepoConfig struct {
	Repo      string `mapstructure:"repo,omitempty" jsonschema:"required"`
	Directory string `mapstructure:"directory,omitempty" jsonschema:"required"`
	Branch    string `mapstructure:"branch,omitempty" jsonschema:"required"`
}

func (c ConnectedRepoConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("repo").Short("repository identifier").Required().
		Long("Identifier of the connected repository configured in the Nuon platform").
		Example("my-repo").
		Example("production-infrastructure").
		Field("directory").Short("directory path").Required().
		Long("Path within the repository to the configuration files").
		Example("terraform").
		Example("infra/terraform").
		Field("branch").Short("Git branch").Required().
		Long("Git branch to checkout and use for deployments").
		Example("main").
		Example("develop").
		Example("production")
}
