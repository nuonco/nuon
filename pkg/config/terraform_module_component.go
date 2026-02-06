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
	TerraformVersion string `mapstructure:"terraform_version" toml:"terraform_version" jsonschema:"required"`

	EnvVarMap      map[string]string        `mapstructure:"env_vars,omitempty" toml:"env_vars,omitempty"`
	VarsMap        map[string]string        `mapstructure:"vars,omitempty" toml:"vars,omitempty"`
	VariablesFiles []TerraformVariablesFile `mapstructure:"var_file,omitempty" toml:"var_file,omitempty"`

	PublicRepo    *PublicRepoConfig    `mapstructure:"public_repo,omitempty" toml:"public_repo,omitempty"`
	ConnectedRepo *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty" toml:"connected_repo,omitempty"`

	DriftSchedule *string `mapstructure:"drift_schedule,omitempty" toml:"drift_schedule,omitempty" features:"template" nuonhash:"omitempty"`

	BuildTimeout  string `mapstructure:"build_timeout,omitempty" toml:"build_timeout,omitempty" features:"template" nuonhash:"omitempty"`
	DeployTimeout string `mapstructure:"deploy_timeout,omitempty" toml:"deploy_timeout,omitempty" features:"template" nuonhash:"omitempty"`

	// deprecated
	Variables []TerraformVariable   `mapstructure:"var,omitempty" toml:"var,omitempty"`
	EnvVars   []EnvironmentVariable `mapstructure:"env_var,omitempty" toml:"env_var,omitempty"`
}

func (t TerraformModuleComponentConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		OneOfGroup("vcs", "public_repo", "connected_repo").
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
		Long("Cron expression for periodic drift detection. If not set, drift detection is disabled. Supports templating").
		Field("build_timeout").Short("build operation timeout").
		Long("Duration string for build operations (e.g., \"30m\", \"1h\").").
		Example("30m").
		Example("1h").
		Field("deploy_timeout").Short("deploy operation timeout").
		Long("Duration string for deploy operations (e.g., \"30m\", \"1h\").").
		Example("30m").
		Example("1h").
		Field("var").Deprecated("use vars map instead").
		Field("env_var").Deprecated("use env_vars map instead")
}

func (t *TerraformModuleComponentConfig) Parse() error {
	if t.PublicRepo.IsEmpty() {
		t.PublicRepo = nil
	}
	if t.ConnectedRepo.IsEmpty() {
		t.ConnectedRepo = nil
	}
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
