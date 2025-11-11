package config

import (
	"github.com/invopop/jsonschema"
)

type TerraformVariablesFile struct {
	Contents string `toml:"contents" mapstructure:"contents,omitempty" features:"get,template"`
}

func (t TerraformVariablesFile) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("contents").Short("variable file contents").
		Long("Contents of a Terraform .tfvars file. Supports Nuon templating and external file sources: HTTP(S) URLs (https://example.com/vars.tfvars), git repositories (git::https://github.com/org/repo//path/to/vars.tfvars), file paths (file:///path/to/vars.tfvars), and relative paths (./vars.tfvars)")
}

// NOTE(jm): components are parsed using mapstructure. Please refer to the wiki entry for more.
type TerraformModuleComponentConfig struct {
	TerraformVersion string `mapstructure:"terraform_version" jsonschema:"required"`

	EnvVarMap      map[string]string        `mapstructure:"env_vars,omitempty"`
	VarsMap        map[string]string        `mapstructure:"vars,omitempty"`
	VariablesFiles []TerraformVariablesFile `mapstructure:"var_file,omitempty"`

	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty" jsonschema:"oneof_required=connected_repo"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty"  jsonschema:"oneof_required=public_repo"`

	DriftSchedule *string `mapstructure:"drift_schedule,omitempty" features:"template" nuonhash:"omitempty"`

	// deprecated
	Variables []TerraformVariable   `mapstructure:"var,omitempty" `
	EnvVars   []EnvironmentVariable `mapstructure:"env_var,omitempty"`
}

func (t TerraformModuleComponentConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("source").Short("source path or URL").
		Long("Optional source path or URL for the component configuration. Supports HTTP(S) URLs, git repositories, file paths, and relative paths (./). Examples: https://example.com/config.yaml, git::https://github.com/org/repo//config.yaml, file:///path/to/config.yaml, ./local/config.yaml").
		Field("type").Short("component type").
		Field("name").Short("component name").
		Field("var_name").Short("variable name for component output").
		Long("Optional name to use when storing component outputs as variables. If not specified, uses the component name").
		Field("dependencies").Short("component dependencies").
		Long("List of other components that must be deployed before this component. Automatically extracted from template references").
		Field("terraform_version").Short("Terraform version").Required().
		Long("Version of Terraform to use for deployments").
		Example("1.5.0").
		Example("1.6.0").
		Example("latest").
		Field("env_vars").Short("environment variables").
		Long("Map of environment variables passed to Terraform as key-value pairs").
		Field("vars").Short("Terraform variables").
		Long("Map of Terraform input variables as key-value pairs. Supports templating").
		Field("var_file").Short("Terraform variable files").
		Long("Array of external Terraform variable files to load. Each file contents support templating and external file sources: HTTP(S) URLs (https://example.com/vars.tfvars), git repositories (git::https://github.com/org/repo//path/to/vars.tfvars), file paths (file:///path/to/vars.tfvars), and relative paths (./vars.tfvars)").
		Field("public_repo").Short("public repository configuration").
		Long("Configuration for a public repository accessible without authentication").
		Field("connected_repo").Short("connected repository configuration").
		Long("Configuration for a private repository connected to the Nuon platform").
		Field("drift_schedule").Short("drift detection schedule").
		Long("Cron expression for periodic drift detection. If not set, drift detection is disabled. Supports templating")
}

func (t *TerraformModuleComponentConfig) Parse() error {
	return nil
}

func (t *TerraformModuleComponentConfig) Validate() error {
	if len(t.Variables) > 0 {
		return ErrConfig{
			Description: "the var array is deprecated, please use vars instead.",
		}
	}
	if len(t.EnvVars) > 0 {
		return ErrConfig{
			Description: "the env_var array is deprecated, please use env_vars instead.",
		}
	}

	return nil
}
