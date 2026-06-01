package config

import (
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/pkg/config/source"
)

type AppSandboxConfig struct {
	Source string `mapstructure:"source,omitempty" toml:"source,omitempty"`

	TerraformVersion string               `mapstructure:"terraform_version" toml:"terraform_version" jsonschema:"required"`
	ConnectedRepo    *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty" toml:"connected_repo,omitempty" jsonschema:"oneof_required=connected_repo"`
	PublicRepo       *PublicRepoConfig    `mapstructure:"public_repo,omitempty" toml:"public_repo,omitempty" jsonschema:"oneof_required=public_repo"`
	DriftSchedule    *string              `mapstructure:"drift_schedule,omitempty" toml:"drift_schedule,omitempty"`

	EnvVarMap      map[string]string        `mapstructure:"env_vars,omitempty" toml:"env_vars,omitempty"`
	VarsMap        map[string]string        `mapstructure:"vars,omitempty" toml:"vars,omitempty"`
	VariablesFiles []TerraformVariablesFile `mapstructure:"var_file,omitempty" toml:"var_file,omitempty"`
	OperationRoles []EntityOperationRole    `mapstructure:"operation_roles,omitempty" toml:"operation_roles,omitempty"`

	MaxAutoRetries               *int  `mapstructure:"max_auto_retries,omitempty" toml:"max_auto_retries,omitempty" nuonhash:"omitempty"`
	SkipNoops                    *bool `mapstructure:"skip_noops,omitempty" toml:"skip_noops,omitempty" nuonhash:"omitempty"`
	AutoApproveOnPoliciesPassing *bool `mapstructure:"auto_approve_on_policies_passing,omitempty" toml:"auto_approve_on_policies_passing,omitempty" nuonhash:"omitempty"`

	References []refs.Ref `mapstructure:"-" jsonschema:"-" nuonhash:"-"`
}

func (a AppSandboxConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("source").Short("external configuration source").
		Long("Path to an external file containing sandbox configuration (YAML, JSON, or TOML)").
		Field("terraform_version").Short("Terraform version").Required().
		Long("Version of Terraform to use for deployments").
		Example("1.5.0").
		Example("1.6.0").
		Example("latest").
		Field("connected_repo").Short("connected repository configuration").
		Long("Configuration for a private repository connected to the Nuon platform").
		Field("public_repo").Short("public repository configuration").
		Long("Configuration for a public repository accessible without authentication").
		Field("drift_schedule").Short("drift detection schedule").
		Long("Cron expression for periodic drift detection. If not set, drift detection is disabled").
		Example("0 2 * * *").
		Example("*/10 * * * *").
		Field("env_vars").Short("environment variables").
		Long("Map of environment variables passed to Terraform as key-value pairs").
		Field("vars").Short("Terraform variables").
		Long("Map of Terraform input variables as key-value pairs. Supports templating").
		Field("var_file").Short("Terraform variable files").
		Long("Array of external Terraform variable files to load. Each file contents support templating and external file sources: HTTP(S) URLs (https://example.com/vars.tfvars), git repositories (git::https://github.com/org/repo//path/to/vars.tfvars), file paths (file:///path/to/vars.tfvars), and relative paths (./vars.tfvars)").
		Field("operation_roles").Short("operation-specific IAM role assignments").
		Long("Map of sandbox operations to IAM role names. Allows using different roles for different operations (provision, deprovision, reprovision). Roles must be defined in the CloudFormation stack deployed to the customer's AWS account. If not specified, default roles are used based on operation type").
		Field("max_auto_retries").Short("maximum automatic retry attempts on sandbox apply failure").
		Long("Maximum number of automatic retry attempts for failed sandbox provision, reprovision, and deprovision applies. Set to 0 to disable auto-retry. Default: 0 (disabled)").
		Default("0").
		Example("3").
		Example("5").
		Field("skip_noops").Short("Skip the sandbox apply when the plan has no changes (a no-op). Defaults to false").
		Default("false").
		Example("true").
		Field("auto_approve_on_policies_passing").Short("Auto-approve the sandbox apply when all policy checks pass. Defaults to false").
		Default("false").
		Example("true")
}

func (a *AppSandboxConfig) parse() error {
	if a == nil {
		return ErrConfig{
			Description: "an app sandbox config is required",
		}
	}

	references, err := refs.Parse(a)
	if err != nil {
		return errors.Wrap(err, "unable to parse components")
	}
	a.References = references

	if a.Source == "" {
		return nil
	}

	obj, err := source.LoadSource(a.Source)
	if err != nil {
		return ErrConfig{
			Description: fmt.Sprintf("unable to load source %s", a.Source),
			Err:         err,
		}
	}

	if err := mapstructure.Decode(obj, &a); err != nil {
		return err
	}

	if a.ConnectedRepo == nil && a.PublicRepo == nil {
		return ErrConfig{
			Description: "either a public repo or connected repo block must be set",
		}
	}
	return nil
}
