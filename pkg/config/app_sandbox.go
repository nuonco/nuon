package config

import (
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/config/refs"
	"github.com/powertoolsdev/mono/pkg/config/source"
)

type AppSandboxConfig struct {
	Source string `mapstructure:"source,omitempty"`

	TerraformVersion string               `mapstructure:"terraform_version" jsonschema:"required"`
	ConnectedRepo    *ConnectedRepoConfig `mapstructure:"connected_repo,omitempty" jsonschema:"oneof_required=connected_repo"`
	PublicRepo       *PublicRepoConfig    `mapstructure:"public_repo,omitempty" jsonschema:"oneof_required=public_repo"`
	DriftSchedule    *string              `mapstructure:"drift_schedule,omitempty"`

	EnvVarMap      map[string]string        `mapstructure:"env_vars,omitempty"`
	VarsMap        map[string]string        `mapstructure:"vars,omitempty"`
	VariablesFiles []TerraformVariablesFile `mapstructure:"var_file,omitempty"`
	References     []refs.Ref               `mapstructure:"-" jsonschema:"-" nuonhash:"-"`
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
		Field("env_vars").Short("environment variables").
		Long("Map of environment variables passed to Terraform as key-value pairs").
		Field("vars").Short("Terraform variables").
		Long("Map of Terraform input variables as key-value pairs. Supports templating").
		Field("var_file").Short("Terraform variable files").
		Long("Array of external Terraform variable files to load. Each file contents support templating and external file sources: HTTP(S) URLs (https://example.com/vars.tfvars), git repositories (git::https://github.com/org/repo//path/to/vars.tfvars), file paths (file:///path/to/vars.tfvars), and relative paths (./vars.tfvars)")
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
