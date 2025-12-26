package config

import (
	"github.com/invopop/jsonschema"
)

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type DockerBuildComponentConfig struct {
	Dockerfile string `mapstructure:"dockerfile" jsonschema:"required"`

	EnvVarMap map[string]string `mapstructure:"env_vars,omitempty"`

	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty" jsonschema:"oneof_required=connected_repo"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty"  jsonschema:"oneof_required=public_repo"`

	// NOTE: the following parameters are not supported in the provider
	// Target	     string		   `mapstructure:"target" toml:"target"`
	// BuildArgs []string		`mapstructure:"build_args" toml:"build_args"`
}

func (d DockerBuildComponentConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("dockerfile").Short("path to the Dockerfile").Required().
		Long("Path to the Dockerfile to build. Supports external file sources: HTTP(S) URLs (https://example.com/Dockerfile), git repositories (git::https://github.com/org/repo//Dockerfile), file paths (file:///path/to/Dockerfile), and relative paths (./Dockerfile)").
		Example("Dockerfile").
		Example("docker/Dockerfile.prod").
		Example("https://github.com/myorg/myrepo/raw/main/Dockerfile").
		Field("env_vars").Short("build environment variables").
		Long("Map of environment variables to pass to the Docker build command. Available during the build process. Supports Go templating").
		Example("GOLANG_VERSION").
		Example("NODE_ENV").
		Field("public_repo").Short("public repository containing Dockerfile").OneOfRequired("repository_source").
		Long("Clone a public GitHub repository containing the Dockerfile and build context. Requires repo, branch, and optionally directory").
		Field("connected_repo").Short("connected repository containing Dockerfile").OneOfRequired("repository_source").
		Long("Use a Nuon-connected repository containing the Dockerfile and build context. Requires repo, branch, and optionally directory")
}

func (t *DockerBuildComponentConfig) Validate() error {
	return nil
}

func (t *DockerBuildComponentConfig) Parse() error {
	return nil
}
